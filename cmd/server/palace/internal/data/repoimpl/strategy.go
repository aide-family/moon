package repoimpl

import (
	"context"
	"fmt"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewStrategyRepository 创建策略仓库
func NewStrategyRepository(data *data.Data) repository.Strategy {
	return &strategyRepositoryImpl{
		data: data,
	}
}

type strategyRepositoryImpl struct {
	data *data.Data
}

func (s *strategyRepositoryImpl) Sync(ctx context.Context, id uint32) error {
	s.syncStrategiesByIds(ctx, id)
	return nil
}

func (s *strategyRepositoryImpl) GetTeamStrategyLevel(ctx context.Context, params *bo.GetTeamStrategyLevelParams) (*bizmodel.StrategyLevel, error) {
	bizQuery, err := getTeamIDBizQuery(s.data, params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}

	return bizQuery.StrategyLevel.WithContext(ctx).Preload(field.Associations).Where(bizQuery.StrategyLevel.ID.Eq(params.LevelID)).First()
}

func (s *strategyRepositoryImpl) GetTeamStrategy(ctx context.Context, params *bo.GetTeamStrategyParams) (*bizmodel.Strategy, error) {
	bizQuery, err := getTeamIDBizQuery(s.data, params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}

	return bizQuery.Strategy.WithContext(ctx).Preload(field.Associations).Where(bizQuery.Strategy.ID.Eq(params.StrategyID)).First()
}

func (s *strategyRepositoryImpl) GetStrategyByIds(ctx context.Context, ids []uint32) ([]*bizmodel.Strategy, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.Strategy.WithContext(ctx).Preload(bizQuery.Strategy.Group).Where(bizQuery.Strategy.ID.In(ids...)).Find()
}

func (s *strategyRepositoryImpl) Eval(ctx context.Context, strategy *bo.Strategy) (*bo.Alarm, error) {
	// TODO 告警评估
	return nil, merr.ErrorNotification("未实现本地告警评估逻辑")
}

func (s *strategyRepositoryImpl) syncStrategiesByIds(ctx context.Context, strategyIds ...uint32) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if err != nil {
		log.Errorw("method", "syncStrategiesByIds", "err", err)
		return
	}
	// 关联查询等级等明细信息
	strategies, err := bizQuery.Strategy.WithContext(ctx).Unscoped().
		Where(bizQuery.Strategy.ID.In(strategyIds...)).
		Preload(field.Associations).
		Preload(bizQuery.Strategy.AlarmNoticeGroups).
		Preload(bizQuery.Strategy.Levels.AlarmGroups).
		Preload(bizQuery.Strategy.Levels.LabelNotices.AlarmGroups).
		Find()
	if !types.IsNil(err) {
		log.Errorw("method", "syncStrategiesByIds", "err", err)
		return
	}
	go func() {
		defer after.RecoverX()
		for _, strategy := range strategies {
			// TODO 完成其他策略类型转换
			items := builder.NewParamsBuild(ctx).StrategyModuleBuilder().DoStrategyBuilder().ToBos(strategy)
			if len(items) == 0 {
				continue
			}
			for _, item := range items {
				if err = s.data.GetStrategyQueue().Push(item.Message()); err != nil {
					return
				}
			}
		}
	}()
}

func (s *strategyRepositoryImpl) UpdateStatus(ctx context.Context, params *bo.UpdateStrategyStatusParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.WithContext(ctx).
		Strategy.Where(bizQuery.Strategy.ID.In(params.Ids...)).
		Update(bizQuery.Strategy.Status, params.Status)
	if !types.IsNil(err) {
		return err
	}
	s.syncStrategiesByIds(ctx, params.Ids...)
	return nil
}

func (s *strategyRepositoryImpl) DeleteByID(ctx context.Context, strategyID uint32) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	strategy := &bizmodel.Strategy{AllFieldModel: model.AllFieldModel{ID: strategyID}}
	defer s.syncStrategiesByIds(ctx, strategyID)
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		// 移除策略数据源中间表关联关系
		if err = tx.Strategy.Datasource.Model(strategy).Clear(); err != nil {
			return err
		}

		// 移除策略类型中间表关联关系
		if err = tx.Strategy.Categories.Model(strategy).Clear(); err != nil {
			return err
		}

		// 移除告警组中间表
		if err = tx.Strategy.AlarmNoticeGroups.Model(strategy).Clear(); err != nil {
			return err
		}

		// 移除策略等级中间表
		if _, err = tx.StrategyLevel.WithContext(ctx).Where(tx.StrategyLevel.StrategyID.Eq(strategyID)).Delete(); !types.IsNil(err) {
			return err
		}

		if _, err = tx.Strategy.WithContext(ctx).Where(tx.Strategy.ID.Eq(strategyID)).Delete(); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

// 校验策略名称是否存在
func (s *strategyRepositoryImpl) checkStrategyName(ctx context.Context, name string, strategyGroupID, id uint32) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	strategyDo, err := bizQuery.Strategy.WithContext(ctx).
		Where(bizQuery.Strategy.Name.Eq(name)).
		Where(bizQuery.Strategy.GroupID.Eq(strategyGroupID)).
		First()
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return nil
	}

	if (id > 0 && strategyDo.ID != id) || id == 0 {
		return merr.ErrorI18nAlertStrategyNameDuplicate(ctx)
	}

	return nil
}

// 检验策略组是否存在
func (s *strategyRepositoryImpl) checkStrategyGroup(ctx context.Context, groupIDs ...uint32) error {
	if len(groupIDs) == 0 {
		return nil
	}
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	// 去重
	groupIDs = types.SliceUnique(groupIDs)
	// 校验告警组是否存在
	count, err := bizQuery.StrategyGroup.WithContext(ctx).
		Where(bizQuery.StrategyGroup.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.StrategyGroup.ID.In(groupIDs...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(count) != len(groupIDs) {
		return merr.ErrorI18nAlertStrategyGroupNotFound(ctx)
	}
	return nil
}

// 检验告警组是否存在
func (s *strategyRepositoryImpl) checkAlarmGroup(ctx context.Context, groupIDs ...uint32) error {
	if len(groupIDs) == 0 {
		return nil
	}
	// 去重
	groupIDs = types.SliceUnique(groupIDs)
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	// 校验告警组是否存在
	count, err := bizQuery.AlarmNoticeGroup.WithContext(ctx).
		Where(bizQuery.AlarmNoticeGroup.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.AlarmNoticeGroup.ID.In(groupIDs...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(count) != len(groupIDs) {
		return merr.ErrorI18nAlertAlertGroupNotFound(ctx)
	}
	return nil
}

// 检验告警等级是否存在
func (s *strategyRepositoryImpl) checkAlarmLevel(ctx context.Context, levelIDs ...uint32) error {
	if len(levelIDs) == 0 {
		return nil
	}
	// 去重
	levelIDs = types.SliceUnique(levelIDs)
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	// 校验告警组是否存在
	count, err := bizQuery.SysDict.WithContext(ctx).
		Where(bizQuery.SysDict.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.SysDict.DictType.Eq(vobj.DictTypeAlarmLevel.GetValue())).
		Where(bizQuery.SysDict.ID.In(levelIDs...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(count) != len(levelIDs) {
		return merr.ErrorI18nAlertAlertLevelNotFound(ctx)
	}
	return nil
}

// 检验数据源是否存在
func (s *strategyRepositoryImpl) checkDataSource(ctx context.Context, dataSourceIds ...uint32) error {
	if len(dataSourceIds) == 0 {
		return nil
	}
	// 去重
	dataSourceIds = types.SliceUnique(dataSourceIds)
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	// 校验数据源是否存在
	count, err := bizQuery.Datasource.WithContext(ctx).
		Where(bizQuery.Datasource.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.Datasource.ID.In(dataSourceIds...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(count) != len(dataSourceIds) {
		return merr.ErrorI18nAlertDatasourceNotFound(ctx)
	}
	return nil
}

// 检验策略类型是否存在
func (s *strategyRepositoryImpl) checkStrategyCategory(ctx context.Context, categoryIds ...uint32) error {
	if len(categoryIds) == 0 {
		return nil
	}
	// 去重
	categoryIds = types.SliceUnique(categoryIds)
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	// 校验策略类型是否存在
	count, err := bizQuery.SysDict.WithContext(ctx).
		Where(bizQuery.SysDict.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.SysDict.DictType.Eq(vobj.DictTypeStrategyCategory.GetValue())).
		Where(bizQuery.SysDict.ID.In(categoryIds...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(count) != len(categoryIds) {
		return merr.ErrorI18nAlertStrategyTypeNotExist(ctx)
	}
	return nil
}

// 校验策略模板是否存在
func (s *strategyRepositoryImpl) checkStrategyTemplate(ctx context.Context, templateID uint32) error {
	if templateID == 0 {
		return nil
	}
	var err error
	sourceType := middleware.GetSourceType(ctx)
	if sourceType.IsTeam() {
		// TODO 查询系统模板是否存在
		mainQuery := query.Use(s.data.GetMainDB(ctx))
		_, err = mainQuery.WithContext(ctx).StrategyTemplate.
			Where(mainQuery.StrategyTemplate.Status.Eq(vobj.StatusEnable.GetValue())).
			Where(mainQuery.StrategyTemplate.ID.Eq(templateID)).First()
	} else {
		// TODO 查询团队模板是否存在
		bizQuery, errX := getBizQuery(ctx, s.data)
		if !types.IsNil(errX) {
			return errX
		}
		_, err = bizQuery.StrategyTemplate.WithContext(ctx).
			Where(bizQuery.StrategyTemplate.Status.Eq(vobj.StatusEnable.GetValue())).
			Where(bizQuery.StrategyTemplate.ID.Eq(templateID)).First()
	}

	if types.IsNil(err) {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorI18nAlertStrategyTemplateNotFound(ctx)
	}
	return err
}

func (s *strategyRepositoryImpl) getAlarmGroupIds(params *bo.CreateStrategyParams) []uint32 {
	alarmGroupIds := params.AlarmGroupIds
	for _, level := range params.Levels {
		alarmGroupIds = append(alarmGroupIds, level.AlarmGroupIds...)
		for _, notice := range level.LabelNotices {
			alarmGroupIds = append(alarmGroupIds, notice.AlarmGroupIds...)
		}
	}
	return alarmGroupIds
}

func (s *strategyRepositoryImpl) getLevelIds(params *bo.CreateStrategyParams) []uint32 {
	return types.SliceTo(params.Levels, func(level *bo.CreateStrategyLevel) uint32 { return level.LevelID })
}

func (s *strategyRepositoryImpl) CreateStrategy(ctx context.Context, params *bo.CreateStrategyParams) (*bizmodel.Strategy, error) {
	if err := s.checkStrategyName(ctx, params.Name, params.GroupID, 0); !types.IsNil(err) {
		return nil, err
	}
	if err := s.checkStrategyGroup(ctx, params.GroupID); !types.IsNil(err) {
		return nil, err
	}

	if err := s.checkAlarmGroup(ctx, s.getAlarmGroupIds(params)...); !types.IsNil(err) {
		return nil, err
	}

	if err := s.checkAlarmLevel(ctx, s.getLevelIds(params)...); !types.IsNil(err) {
		return nil, err
	}

	if err := s.checkDataSource(ctx, params.DatasourceIDs...); !types.IsNil(err) {
		return nil, err
	}

	if err := s.checkStrategyCategory(ctx, params.CategoriesIds...); !types.IsNil(err) {
		return nil, err
	}

	if err := s.checkStrategyTemplate(ctx, params.TemplateID); !types.IsNil(err) {
		return nil, err
	}

	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyModel := createStrategyParamsToModel(ctx, params)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.Strategy.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(strategyModel); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	s.syncStrategiesByIds(ctx, strategyModel.ID)
	return strategyModel, nil
}

func (s *strategyRepositoryImpl) UpdateByID(ctx context.Context, params *bo.UpdateStrategyParams) error {
	updateParam := params.UpdateParam
	if updateParam == nil {
		panic("strategyRepo UpdateByID method params UpdateParam field is nil")
	}
	if err := s.checkStrategyName(ctx, updateParam.Name, updateParam.GroupID, params.ID); !types.IsNil(err) {
		return err
	}
	if err := s.checkStrategyGroup(ctx, updateParam.GroupID); !types.IsNil(err) {
		return err
	}

	if err := s.checkAlarmGroup(ctx, s.getAlarmGroupIds(updateParam)...); !types.IsNil(err) {
		return err
	}

	if err := s.checkAlarmLevel(ctx, s.getLevelIds(updateParam)...); !types.IsNil(err) {
		return err
	}

	if err := s.checkDataSource(ctx, updateParam.DatasourceIDs...); !types.IsNil(err) {
		return err
	}

	if err := s.checkStrategyCategory(ctx, updateParam.CategoriesIds...); !types.IsNil(err) {
		return err
	}

	if err := s.checkStrategyTemplate(ctx, updateParam.TemplateID); !types.IsNil(err) {
		return err
	}

	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}

	datasource := types.SliceToWithFilter(updateParam.DatasourceIDs, func(datasourceId uint32) (*bizmodel.Datasource, bool) {
		if datasourceId <= 0 {
			return nil, false
		}
		return &bizmodel.Datasource{
			AllFieldModel: model.AllFieldModel{ID: datasourceId},
		}, true
	})
	// 策略类型
	categories := types.SliceToWithFilter(updateParam.CategoriesIds, func(categoriesID uint32) (*bizmodel.SysDict, bool) {
		if categoriesID <= 0 {
			return nil, false
		}
		return &bizmodel.SysDict{
			AllFieldModel: model.AllFieldModel{ID: categoriesID},
		}, true
	})

	// 告警分组
	alarmGroups := types.SliceToWithFilter(updateParam.AlarmGroupIds, func(alarmGroupsID uint32) (*bizmodel.AlarmNoticeGroup, bool) {
		if alarmGroupsID <= 0 {
			return nil, false
		}
		return &bizmodel.AlarmNoticeGroup{AllFieldModel: model.AllFieldModel{ID: alarmGroupsID}}, true
	})

	strategyLevels := createStrategyLevelParamsToModel(ctx, updateParam.Levels)
	strategyLevels = types.SliceToWithFilter(strategyLevels, func(item *bizmodel.StrategyLevel) (*bizmodel.StrategyLevel, bool) {
		item.StrategyID = params.ID
		return item, true
	})

	strategyModel := &bizmodel.Strategy{AllFieldModel: model.AllFieldModel{ID: params.ID}}
	defer s.syncStrategiesByIds(ctx, params.ID)
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		// Datasource
		if err = tx.Strategy.Datasource.
			Model(strategyModel).Replace(datasource...); !types.IsNil(err) {
			return err
		}
		// Categories
		if err = tx.Strategy.Categories.
			Model(strategyModel).Replace(categories...); !types.IsNil(err) {
			return err
		}
		// AlarmGroups
		if err = tx.Strategy.AlarmNoticeGroups.Model(strategyModel).Replace(alarmGroups...); !types.IsNil(err) {
			return err
		}
		// StrategyLevel
		// 删除策略等级数据
		if _, err = tx.StrategyLevel.WithContext(ctx).Where(tx.StrategyLevel.StrategyID.Eq(params.ID)).Delete(); !types.IsNil(err) {
			return err
		}

		if err = tx.StrategyLevel.Clauses(clause.OnConflict{UpdateAll: true}).Create(strategyLevels...); !types.IsNil(err) {
			return err
		}

		// 更新策略
		if _, err = tx.Strategy.WithContext(ctx).Where(tx.Strategy.ID.Eq(params.ID)).UpdateSimple(
			tx.Strategy.Name.Value(updateParam.Name),
			tx.Strategy.Step.Value(updateParam.Step),
			tx.Strategy.Remark.Value(updateParam.Remark),
			tx.Strategy.Expr.Value(updateParam.Expr),
			tx.Strategy.Labels.Value(updateParam.Labels),
			tx.Strategy.Annotations.Value(updateParam.Annotations),
			tx.Strategy.GroupID.Value(updateParam.GroupID),
		); !types.IsNil(err) {
			return err
		}

		return nil
	})
}

func (s *strategyRepositoryImpl) GetByID(ctx context.Context, strategyID uint32) (*bizmodel.Strategy, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.Strategy.WithContext(ctx)
	strategy, err := bizWrapper.
		Where(bizQuery.Strategy.ID.Eq(strategyID)).
		Preload(field.Associations).
		Preload(bizQuery.Strategy.Levels.AlarmPage).
		Preload(bizQuery.Strategy.Levels.AlarmGroups).
		Preload(bizQuery.Strategy.Levels.LabelNotices).
		Preload(bizQuery.Strategy.Levels.LabelNotices.AlarmGroups).
		Preload(bizQuery.Strategy.Levels.Level).
		First()
	if !types.IsNil(err) {
		return nil, err
	}
	return strategy, nil
}

func (s *strategyRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryStrategyListParams) ([]*bizmodel.Strategy, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyWrapper := bizQuery.Strategy.WithContext(ctx)

	var wheres []gen.Condition
	if !types.TextIsNull(params.Alert) {
		wheres = append(wheres, bizQuery.Strategy.Name.Like(params.Alert))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.Strategy.Status.Eq(params.Status.GetValue()))
	}

	if !types.TextIsNull(params.Keyword) {
		strategyWrapper = strategyWrapper.Or(bizQuery.Strategy.Name.Like(params.Keyword))
		strategyWrapper = strategyWrapper.Or(bizQuery.Strategy.Remark.Like(params.Keyword))

		dictWrapper := query.Use(s.data.GetMainDB(ctx)).SysDict.WithContext(ctx)

		dictWrapper = dictWrapper.Or(bizQuery.SysDict.Name.Like(params.Keyword))
		dictWrapper = dictWrapper.Or(bizQuery.SysDict.Value.Like(params.Keyword))
		dictWrapper = dictWrapper.Or(bizQuery.SysDict.Remark.Like(params.Keyword))

		sysDicts, err := dictWrapper.Find()
		if err != nil {
			return nil, err
		}

		categoriesIds := types.SliceTo(sysDicts, func(item *model.SysDict) uint32 {
			return item.ID
		})

		var strategyTemplateIds []uint32
		mainQuery := query.Use(s.data.GetMainDB(ctx))
		strategyTemplateCategories := mainQuery.StrategyTemplateCategories.WithContext(ctx)
		_ = strategyTemplateCategories.Where(mainQuery.StrategyTemplateCategories.SysDictID.In(categoriesIds...)).
			Select(mainQuery.StrategyTemplateCategories.StrategyTemplateID).
			Scan(&strategyTemplateIds)
		if len(strategyTemplateIds) > 0 {
			strategyWrapper = strategyWrapper.Or(bizQuery.Strategy.TemplateID.In(strategyTemplateIds...))
		}
	}

	strategyWrapper = strategyWrapper.Where(wheres...).Preload(field.Associations)

	if strategyWrapper, err = types.WithPageQuery(strategyWrapper, params.Page); err != nil {
		return nil, err
	}

	return strategyWrapper.Order(bizQuery.Strategy.ID.Desc()).Find()
}

func (s *strategyRepositoryImpl) CopyStrategy(ctx context.Context, strategyID uint32) (*bizmodel.Strategy, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyWrapper := bizQuery.Strategy.WithContext(ctx)
	strategy, err := strategyWrapper.Where(bizQuery.Strategy.ID.Eq(strategyID)).Preload(field.Associations).First()
	if !types.IsNil(err) {
		return nil, err
	}
	strategy.Name = fmt.Sprintf("%s-%d-%s", strategy.Name, strategyID, "copy")
	strategy.ID = 0
	strategy.Status = vobj.StatusDisable
	copyLevels := make([]*bizmodel.StrategyLevel, 0, len(strategy.Levels))
	for _, level := range strategy.Levels {
		level.ID = 0
		copyLevels = append(copyLevels, level)
	}
	strategy.Levels = copyLevels
	if err := strategyWrapper.WithContext(ctx).Create(strategy); !types.IsNil(err) {
		return nil, err
	}

	s.syncStrategiesByIds(ctx, strategy.ID)
	return strategy, nil
}

func createStrategyLevelParamsToModel(ctx context.Context, params []*bo.CreateStrategyLevel) []*bizmodel.StrategyLevel {
	strategyLevel := types.SliceTo(params, func(item *bo.CreateStrategyLevel) *bizmodel.StrategyLevel {
		templateLevel := &bizmodel.StrategyLevel{
			AllFieldModel: model.AllFieldModel{ID: item.ID},
			StrategyID:    item.StrategyID,
			Duration:      item.Duration,
			Count:         item.Count,
			SustainType:   item.SustainType,
			Interval:      item.Interval,
			Condition:     item.Condition,
			Threshold:     item.Threshold,
			LevelID:       item.LevelID,
			Status:        item.Status,
			AlarmPage: types.SliceTo(item.AlarmPageIds, func(pageID uint32) *bizmodel.SysDict {
				return &bizmodel.SysDict{
					AllFieldModel: model.AllFieldModel{
						ID: pageID,
					},
				}
			}),
			AlarmGroups: types.SliceTo(item.AlarmGroupIds, func(groupID uint32) *bizmodel.AlarmNoticeGroup {
				return &bizmodel.AlarmNoticeGroup{AllFieldModel: model.AllFieldModel{ID: groupID}}
			}),
			LabelNotices: types.SliceTo(item.LabelNotices, func(notice *bo.StrategyLabelNotice) *bizmodel.StrategyLabelNotice {
				return &bizmodel.StrategyLabelNotice{
					Name:    notice.Name,
					Value:   notice.Value,
					LevelID: item.LevelID,
					AlarmGroups: types.SliceTo(notice.AlarmGroupIds, func(groupID uint32) *bizmodel.AlarmNoticeGroup {
						return &bizmodel.AlarmNoticeGroup{AllFieldModel: model.AllFieldModel{ID: groupID}}
					}),
				}
			}),
		}
		templateLevel.WithContext(ctx)
		return templateLevel
	})
	return strategyLevel
}

func createStrategyParamsToModel(ctx context.Context, params *bo.CreateStrategyParams) *bizmodel.Strategy {
	strategyModel := &bizmodel.Strategy{
		TemplateID:     params.TemplateID,
		GroupID:        params.GroupID,
		TemplateSource: params.TemplateSource,
		Name:           params.Name,
		Expr:           params.Expr,
		Labels:         params.Labels,
		Annotations:    params.Annotations,
		Remark:         params.Remark,
		Status:         vobj.Status(params.Status.GetValue()),
		Step:           params.Step,
		Datasource: types.SliceToWithFilter(params.DatasourceIDs, func(datasourceId uint32) (*bizmodel.Datasource, bool) {
			if datasourceId <= 0 {
				return nil, false
			}
			return &bizmodel.Datasource{
				AllFieldModel: model.AllFieldModel{ID: datasourceId},
			}, true
		}),
		Categories: types.SliceToWithFilter(params.CategoriesIds, func(categoriesID uint32) (*bizmodel.SysDict, bool) {
			if categoriesID <= 0 {
				return nil, false
			}
			return &bizmodel.SysDict{
				AllFieldModel: model.AllFieldModel{ID: categoriesID},
			}, true
		}),
		Levels: createStrategyLevelParamsToModel(ctx, params.Levels),
		AlarmNoticeGroups: types.SliceToWithFilter(params.AlarmGroupIds, func(groupID uint32) (*bizmodel.AlarmNoticeGroup, bool) {
			return &bizmodel.AlarmNoticeGroup{
				AllFieldModel: model.AllFieldModel{
					ID: groupID,
				},
			}, true
		}),
	}
	strategyModel.WithContext(ctx)
	return strategyModel
}

func createStrategyLabelsToModel(ctx context.Context, labels []*bo.StrategyLabelNotice) []*bizmodel.StrategyLabelNotice {
	strategyLabelsModels := types.SliceTo(labels, func(label *bo.StrategyLabelNotice) *bizmodel.StrategyLabelNotice {
		labelsModel := &bizmodel.StrategyLabelNotice{
			Name:  label.Name,
			Value: label.Value,
			AlarmGroups: types.SliceToWithFilter(label.AlarmGroupIds, func(groupID uint32) (*bizmodel.AlarmNoticeGroup, bool) {
				if groupID <= 0 {
					return nil, false
				}
				return &bizmodel.AlarmNoticeGroup{
					AllFieldModel: model.AllFieldModel{
						ID: groupID,
					},
				}, true
			}),
		}
		labelsModel.WithContext(ctx)
		return labelsModel
	})
	return strategyLabelsModels
}
