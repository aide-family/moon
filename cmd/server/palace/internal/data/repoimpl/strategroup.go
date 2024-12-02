package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// NewStrategyGroupRepository 创建策略分组仓库
func NewStrategyGroupRepository(data *data.Data, strategyRepo repository.Strategy) repository.StrategyGroup {
	return &strategyGroupRepositoryImpl{
		data:         data,
		strategyRepo: strategyRepo,
	}
}

// NewStrategyCountRepository 创建策略计数仓库
func NewStrategyCountRepository(data *data.Data) repository.StrategyCountRepo {
	return &strategyCountRepositoryImpl{
		data: data,
	}
}

type (
	strategyGroupRepositoryImpl struct {
		data         *data.Data
		strategyRepo repository.Strategy
	}

	strategyCountRepositoryImpl struct {
		data *data.Data
	}
)

func (s *strategyGroupRepositoryImpl) syncStrategiesByGroupIds(ctx context.Context, groupIds ...uint32) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return
	}

	for _, groupID := range groupIds {
		strategies, err := bizQuery.Strategy.WithContext(ctx).Unscoped().
			Where(bizQuery.Strategy.GroupID.Eq(groupID)).
			Preload(field.Associations).
			Find()
		if !types.IsNil(err) {
			continue
		}

		strategyIds := types.To(strategies, func(strategy *bizmodel.Strategy) uint32 {
			return strategy.ID
		})
		metricLevels, err := s.strategyRepo.GetStrategyMetricLevels(ctx, strategyIds)
		if err != nil {
			return
		}

		strategyMQLevels, err := s.strategyRepo.GetStrategyMQLevels(ctx, strategyIds)
		if err != nil {
			return
		}

		metricsLevelMap := types.ToMapSlice(metricLevels, func(level *bizmodel.StrategyMetricsLevel) uint32 {
			return level.StrategyID
		})

		mqLevelMap := types.ToMapSlice(strategyMQLevels, func(level *bizmodel.StrategyMQLevel) uint32 {
			return level.StrategyID
		})

		strategyDetailMap := &bo.StrategyLevelDetailModel{MetricsLevelMap: metricsLevelMap, MQLevelMap: mqLevelMap}
		go func() {
			defer after.RecoverX()
			for _, strategy := range strategies {
				items := builder.NewParamsBuild(ctx).StrategyModuleBuilder().DoStrategyBuilder().WithStrategyLevelDetail(strategyDetailMap).ToBos(strategy)
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
}

func (s *strategyCountRepositoryImpl) FindStrategyCount(ctx context.Context, params *bo.GetStrategyCountParams) ([]*bo.StrategyCountModel, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyGroupIds := params.StrategyGroupIds
	var totals []*bo.StrategyCountModel
	wheres := make([]gen.Condition, 0, 2)
	if len(params.StrategyGroupIds) > 0 {
		wheres = append(wheres, bizQuery.Strategy.GroupID.In(strategyGroupIds...))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.Strategy.Status.Eq(params.Status.GetValue()))
	}
	err = bizQuery.Strategy.WithContext(ctx).Where(wheres...).
		Select(bizQuery.Strategy.GroupID.Count().As("total"), bizQuery.Strategy.GroupID.As("group_id")).
		Group(bizQuery.Strategy.GroupID).Scan(&totals)
	if !types.IsNil(err) {
		return nil, err
	}
	return totals, nil
}

// 检查策略组名称是否存在
func (s *strategyGroupRepositoryImpl) checkStrategyGroupName(ctx context.Context, name string, id uint32) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	strategyGroupDo, err := bizQuery.StrategyGroup.WithContext(ctx).Where(bizQuery.StrategyGroup.Name.Eq(name)).First()
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if (id > 0 && strategyGroupDo.ID != id) || id == 0 {
		return merr.ErrorI18nAlertStrategyGroupNameDuplicate(ctx)
	}
	return nil
}

// 检查策略类型是否存在
func (s *strategyGroupRepositoryImpl) checkStrategyGroupCategories(ctx context.Context, categoriesIds []uint32) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}

	count, err := bizQuery.SysDict.WithContext(ctx).
		Where(bizQuery.SysDict.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.SysDict.DictType.Eq(vobj.DictTypeStrategyGroupCategory.GetValue())).
		Where(bizQuery.SysDict.ID.In(categoriesIds...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(count) != len(categoriesIds) {
		return merr.ErrorI18nAlertStrategyGroupTypeNotExist(ctx)
	}
	return nil
}

func (s *strategyGroupRepositoryImpl) CreateStrategyGroup(ctx context.Context, params *bo.CreateStrategyGroupParams) (*bizmodel.StrategyGroup, error) {
	if err := s.checkStrategyGroupName(ctx, params.Name, 0); !types.IsNil(err) {
		return nil, err
	}
	if err := s.checkStrategyGroupCategories(ctx, params.CategoriesIds); !types.IsNil(err) {
		return nil, err
	}
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyGroupModel := createStrategyGroupParamsToModel(ctx, params)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		return tx.StrategyGroup.WithContext(ctx).Create(strategyGroupModel)
	})
	if !types.IsNil(err) {
		return nil, err
	}
	s.syncStrategiesByGroupIds(ctx, strategyGroupModel.ID)
	return strategyGroupModel, err
}

func (s *strategyGroupRepositoryImpl) UpdateStrategyGroup(ctx context.Context, params *bo.UpdateStrategyGroupParams) error {
	if params.UpdateParam == nil {
		panic("UpdateStrategyGroup method params UpdateParam field is nil")
	}
	if _, err := s.GetStrategyGroup(ctx, params.ID); !types.IsNil(err) {
		return err
	}
	if err := s.checkStrategyGroupName(ctx, params.UpdateParam.Name, params.ID); !types.IsNil(err) {
		return err
	}
	if err := s.checkStrategyGroupCategories(ctx, params.UpdateParam.CategoriesIds); !types.IsNil(err) {
		return err
	}
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	defer s.syncStrategiesByGroupIds(ctx, params.ID)
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if !types.IsNil(params.UpdateParam.CategoriesIds) {
			// 更新类型
			Categories := types.SliceToWithFilter(params.UpdateParam.CategoriesIds, func(id uint32) (*bizmodel.SysDict, bool) {
				if id <= 0 {
					return nil, false
				}
				return &bizmodel.SysDict{
					AllFieldModel: model.AllFieldModel{ID: id},
				}, true
			})
			if err = tx.StrategyGroup.Categories.
				Model(&bizmodel.StrategyGroup{AllFieldModel: model.AllFieldModel{ID: params.ID}}).Replace(Categories...); !types.IsNil(err) {
				return err
			}
		}
		// 更新策略分组
		if _, err = tx.StrategyGroup.WithContext(ctx).Where(tx.StrategyGroup.ID.Eq(params.ID)).UpdateSimple(
			tx.StrategyGroup.Name.Value(params.UpdateParam.Name),
			tx.StrategyGroup.Remark.Value(params.UpdateParam.Remark),
		); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (s *strategyGroupRepositoryImpl) DeleteStrategyGroup(ctx context.Context, params *bo.DelStrategyGroupParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	groupModel := &bizmodel.StrategyGroup{AllFieldModel: model.AllFieldModel{ID: params.ID}}
	defer s.syncStrategiesByGroupIds(ctx, params.ID)
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		// 清除策略类型中间表信息
		if err := tx.StrategyGroup.Categories.Model(groupModel).Clear(); err != nil {
			return err
		}

		if _, err = tx.StrategyGroup.WithContext(ctx).Where(tx.StrategyGroup.ID.Eq(params.ID)).Delete(); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (s *strategyGroupRepositoryImpl) GetStrategyGroup(ctx context.Context, groupID uint32) (*bizmodel.StrategyGroup, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyGroupDo, err := bizQuery.StrategyGroup.WithContext(ctx).Where(bizQuery.StrategyGroup.ID.Eq(groupID)).Preload(field.Associations).First()
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastStrategyGroupNotFound(ctx)
		}
		return nil, err
	}
	return strategyGroupDo, err
}

func (s *strategyGroupRepositoryImpl) StrategyGroupPage(ctx context.Context, params *bo.QueryStrategyGroupListParams) ([]*bizmodel.StrategyGroup, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.StrategyGroup.WithContext(ctx)

	var wheres []gen.Condition
	if !types.TextIsNull(params.Name) {
		wheres = append(wheres, bizQuery.StrategyGroup.Name.Like(params.Name))
	}

	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.StrategyGroup.Status.Eq(params.Status.GetValue()))
	}
	if !types.TextIsNull(params.Keyword) {
		bizWrapper = bizWrapper.Or(bizQuery.StrategyGroup.Name.Like(params.Keyword))
		bizWrapper = bizWrapper.Or(bizQuery.StrategyGroup.Remark.Like(params.Keyword))
	}

	// 通过策略分组类型进行查询
	if len(params.CategoriesIds) > 0 {
		var strategyGroupIds []uint32
		_ = bizQuery.StrategyGroupCategories.
			Where(bizQuery.StrategyGroupCategories.SysDictID.In(params.CategoriesIds...)).
			Select(bizQuery.StrategyGroupCategories.StrategyGroupID).
			Scan(&strategyGroupIds)
		if len(strategyGroupIds) > 0 {
			bizWrapper = bizWrapper.Or(bizQuery.StrategyGroup.ID.In(strategyGroupIds...))
		}
	}

	bizWrapper = bizWrapper.Where(wheres...).Preload(bizQuery.StrategyGroup.Categories)

	if bizWrapper, err = types.WithPageQuery(bizWrapper, params.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(bizQuery.StrategyGroup.ID.Desc()).Find()
}

func (s *strategyGroupRepositoryImpl) UpdateStatus(ctx context.Context, params *bo.UpdateStrategyGroupStatusParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	// 关闭当前分组所有的策略
	if params.Status.IsDisable() {
		if _, err := bizQuery.Strategy.WithContext(ctx).Where(bizQuery.Strategy.GroupID.In(params.IDs...)).Update(bizQuery.Strategy.Status, params.Status); err != nil {
			return err
		}
	}

	if _, err = bizQuery.StrategyGroup.WithContext(ctx).Where(bizQuery.StrategyGroup.ID.In(params.IDs...)).Update(bizQuery.StrategyGroup.Status, params.Status); !types.IsNil(err) {
		return err
	}
	s.syncStrategiesByGroupIds(ctx, params.IDs...)
	return nil
}

func createStrategyGroupParamsToModel(ctx context.Context, params *bo.CreateStrategyGroupParams) *bizmodel.StrategyGroup {
	strategyGroup := &bizmodel.StrategyGroup{
		Name:   params.Name,
		Status: params.Status,
		Remark: params.Remark,
		Categories: types.SliceToWithFilter(params.CategoriesIds, func(id uint32) (*bizmodel.SysDict, bool) {
			if id <= 0 {
				return nil, false
			}
			return &bizmodel.SysDict{
				AllFieldModel: model.AllFieldModel{ID: id},
			}, true
		}),
	}
	strategyGroup.WithContext(ctx)
	return strategyGroup
}
