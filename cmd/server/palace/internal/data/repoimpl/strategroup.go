package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

// NewStrategyGroupRepository 创建策略分组仓库
func NewStrategyGroupRepository(data *data.Data) repository.StrategyGroup {
	return &strategyGroupRepositoryImpl{
		data: data,
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
		data *data.Data
	}

	strategyCountRepositoryImpl struct {
		data *data.Data
	}
)

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

func (s *strategyGroupRepositoryImpl) CreateStrategyGroup(ctx context.Context, params *bo.CreateStrategyGroupParams) (*bizmodel.StrategyGroup, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyGroupModel := createStrategyGroupParamsToModel(ctx, params)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.StrategyGroup.WithContext(ctx).Create(strategyGroupModel); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	return strategyGroupModel, err
}

func (s *strategyGroupRepositoryImpl) UpdateStrategyGroup(ctx context.Context, params *bo.UpdateStrategyGroupParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
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
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if _, err = bizQuery.StrategyGroup.WithContext(ctx).Where(bizQuery.StrategyGroup.ID.Eq(params.ID)).Delete(); !types.IsNil(err) {
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
	bizWrapper := bizQuery.StrategyGroup.WithContext(ctx)
	return bizWrapper.Where(bizQuery.StrategyGroup.ID.Eq(groupID)).Preload(field.Associations).First()
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

	bizWrapper = bizWrapper.Where(wheres...).Preload(field.Associations)

	if err := types.WithPageQuery[bizquery.IStrategyGroupDo](bizWrapper, params.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(bizQuery.StrategyGroup.ID.Desc()).Find()
}

func (s *strategyGroupRepositoryImpl) UpdateStatus(ctx context.Context, params *bo.UpdateStrategyGroupStatusParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	bizWrapper := bizQuery.StrategyGroup.WithContext(ctx)
	if _, err = bizWrapper.Where(bizQuery.StrategyGroup.ID.In(params.IDs...)).Update(bizQuery.StrategyGroup.Status, params.Status); !types.IsNil(err) {
		return err
	}
	return nil
}

func createStrategyGroupParamsToModel(ctx context.Context, params *bo.CreateStrategyGroupParams) *bizmodel.StrategyGroup {
	strategyGroup := &bizmodel.StrategyGroup{
		Name:   params.Name,
		Status: vobj.Status(params.Status),
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
