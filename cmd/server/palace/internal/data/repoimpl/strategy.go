package repoimpl

import (
	"context"
	"fmt"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"

	"gorm.io/gen"
	"gorm.io/gen/field"
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
		Find()
	if !types.IsNil(err) {
		log.Errorw("method", "syncStrategiesByIds", "err", err)
		return
	}
	go func() {
		defer after.RecoverX()
		for _, strategy := range strategies {
			items := build.NewBuilder().WithAPIStrategy(strategy).ToBos()
			if items == nil || len(items) == 0 {
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
		if err = tx.Strategy.AlarmGroups.Model(strategy).Clear(); err != nil {
			return err
		}

		// 移除策略等级中间表
		if err = tx.Strategy.StrategyLevel.WithContext(ctx).Model(strategy).Clear(); !types.IsNil(err) {
			return err
		}

		if _, err = tx.Strategy.WithContext(ctx).Where(tx.Strategy.ID.Eq(strategyID)).Delete(); !types.IsNil(err) {
			return err
		}
		s.syncStrategiesByIds(ctx, strategyID)
		return nil
	})
}

func (s *strategyRepositoryImpl) CreateStrategy(ctx context.Context, params *bo.CreateStrategyParams) (*bizmodel.Strategy, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyModel := createStrategyParamsToModel(ctx, params)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.Strategy.WithContext(ctx).Create(strategyModel); !types.IsNil(err) {
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
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	updateParam := params.UpdateParam
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
	alarmGroups := types.SliceToWithFilter(updateParam.AlarmGroupIds, func(alarmGroupsID uint32) (*bizmodel.AlarmGroup, bool) {
		if alarmGroupsID <= 0 {
			return nil, false
		}
		return &bizmodel.AlarmGroup{AllFieldModel: model.AllFieldModel{ID: alarmGroupsID}}, true
	})

	strategyLabelsModels := createStrategyLabelsToModel(ctx, updateParam.StrategyLabels)

	strategyModel := &bizmodel.Strategy{AllFieldModel: model.AllFieldModel{ID: params.ID}}

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
		if err = tx.Strategy.AlarmGroups.Model(strategyModel).Replace(alarmGroups...); !types.IsNil(err) {
			return err
		}
		// StrategyLabels
		if err = tx.Strategy.StrategyNoticeLabels.Model(strategyModel).Replace(strategyLabelsModels...); !types.IsNil(err) {
			return err
		}
		// StrategyLevel
		// 删除策略等级数据
		if _, err = tx.StrategyLevel.WithContext(ctx).Where(tx.StrategyLevel.StrategyID.Eq(params.ID)).Delete(); !types.IsNil(err) {
			return err
		}
		strategyLevels := createStrategyLevelParamsToModel(ctx, updateParam.StrategyLevel)
		if err = tx.StrategyLevel.Create(strategyLevels...); !types.IsNil(err) {
			return err
		}

		// 更新策略
		if _, err = tx.Strategy.WithContext(ctx).Where(tx.Strategy.ID.Eq(params.ID)).UpdateSimple(
			tx.Strategy.Name.Value(updateParam.Name),
			tx.Strategy.Step.Value(updateParam.Step),
			tx.Strategy.Remark.Value(updateParam.Remark),
			tx.Strategy.Expr.Value(updateParam.Expr),
			tx.Strategy.Labels.Value(updateParam.Labels),
			tx.Strategy.GroupID.Value(updateParam.GroupID),
		); !types.IsNil(err) {
			return err
		}
		s.syncStrategiesByIds(ctx, params.ID)
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
		Preload(bizQuery.Strategy.StrategyLevel.AlarmPage).
		Preload(bizQuery.Strategy.StrategyLevel.AlarmGroups).
		Preload(bizQuery.Strategy.StrategyNoticeLabels.RelationField).
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
			strategyWrapper = strategyWrapper.Or(bizQuery.Strategy.StrategyTemplateID.In(strategyTemplateIds...))
		}
	}

	strategyWrapper = strategyWrapper.Where(wheres...).Preload(field.Associations)

	if err := types.WithPageQuery[bizquery.IStrategyDo](strategyWrapper, params.Page); err != nil {
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
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.Strategy.WithContext(ctx).Create(strategy); !types.IsNil(err) {
			return err
		}
		copyLevels := make([]*bizmodel.StrategyLevel, 0, len(strategy.StrategyLevel))
		for _, level := range strategy.StrategyLevel {
			level.StrategyID = strategy.ID
			level.ID = 0
			copyLevels = append(copyLevels, level)
		}
		if err := tx.StrategyLevel.WithContext(ctx).Create(copyLevels...); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	s.syncStrategiesByIds(ctx, strategy.ID)
	return strategy, nil
}

func createStrategyLevelParamsToModel(ctx context.Context, params []*bo.CreateStrategyLevel) []*bizmodel.StrategyLevel {
	strategyLevel := types.SliceTo(params, func(item *bo.CreateStrategyLevel) *bizmodel.StrategyLevel {
		templateLevel := &bizmodel.StrategyLevel{
			StrategyID:  item.StrategyID,
			Duration:    item.Duration,
			Count:       item.Count,
			SustainType: item.SustainType,
			Interval:    item.Interval,
			Condition:   item.Condition,
			Threshold:   item.Threshold,
			LevelID:     item.LevelID,
			Status:      item.Status,
			AlarmPage: types.SliceTo(item.AlarmPageIds, func(pageID uint32) *bizmodel.SysDict {
				return &bizmodel.SysDict{
					AllFieldModel: model.AllFieldModel{
						ID: pageID,
					},
				}
			}),
			AlarmGroups: types.SliceTo(item.AlarmGroupIds, func(groupID uint32) *bizmodel.AlarmGroup {
				return &bizmodel.AlarmGroup{AllFieldModel: model.AllFieldModel{ID: groupID}}
			}),
		}
		templateLevel.WithContext(ctx)
		return templateLevel
	})
	return strategyLevel
}

func createStrategyParamsToModel(ctx context.Context, params *bo.CreateStrategyParams) *bizmodel.Strategy {
	strategyModel := &bizmodel.Strategy{
		Name:                   params.Name,
		StrategyTemplateID:     params.TemplateID,
		StrategyTemplateSource: vobj.StrategyTemplateSource(params.SourceType),
		Expr:                   params.Expr,
		Labels:                 params.Labels,
		Annotations:            params.Annotations,
		Remark:                 params.Remark,
		Status:                 vobj.Status(params.Status.GetValue()),
		Step:                   params.Step,
		GroupID:                params.GroupID,
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
		AlarmGroups: types.SliceToWithFilter(params.AlarmGroupIds, func(groupID uint32) (*bizmodel.AlarmGroup, bool) {
			return &bizmodel.AlarmGroup{
				AllFieldModel: model.AllFieldModel{
					ID: groupID,
				},
			}, true
		}),
		StrategyLevel:        createStrategyLevelParamsToModel(ctx, params.StrategyLevel),
		StrategyNoticeLabels: createStrategyLabelsToModel(ctx, params.StrategyLabels),
	}
	strategyModel.WithContext(ctx)
	return strategyModel
}

func createStrategyLabelsToModel(ctx context.Context, labels []*bo.StrategyLabels) []*bizmodel.StrategyLabels {
	strategyLabelsModels := types.SliceTo(labels, func(label *bo.StrategyLabels) *bizmodel.StrategyLabels {
		labelsModel := &bizmodel.StrategyLabels{
			Name:  label.Name,
			Value: label.Value,
			AlarmGroups: types.SliceToWithFilter(label.AlarmGroupIds, func(groupID uint32) (*bizmodel.AlarmGroup, bool) {
				if groupID <= 0 {
					return nil, false
				}
				return &bizmodel.AlarmGroup{
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
