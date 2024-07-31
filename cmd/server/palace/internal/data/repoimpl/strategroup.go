package repoimpl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewStrategyGroupRepository 创建策略分组仓库
func NewStrategyGroupRepository(data *data.Data) repository.StrategyGroup {
	return &strategyGroupRepositoryImpl{
		data: data,
	}
}

type strategyGroupRepositoryImpl struct {
	data *data.Data
}

func (s strategyGroupRepositoryImpl) CreateStrategyGroup(ctx context.Context, params *bo.CreateStrategyGroupParams) (*bizmodel.StrategyGroup, error) {
	bizDB, err := s.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}
	strategyGroupModel := createStrategyGroupParamsToModel(ctx, params)
	bizQuery := bizquery.Use(bizDB)
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

func (s strategyGroupRepositoryImpl) UpdateStrategyGroup(ctx context.Context, params *bo.UpdateStrategyGroupParams) error {
	bizDB, err := s.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return err
	}
	queryWrapper := bizquery.Use(bizDB)
	return queryWrapper.Transaction(func(tx *bizquery.Query) error {
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

func (s strategyGroupRepositoryImpl) DeleteStrategyGroup(ctx context.Context, params *bo.DelStrategyGroupParams) error {
	bizDB, err := s.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return err
	}
	queryWrapper := bizquery.Use(bizDB)
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if _, err = queryWrapper.StrategyGroup.WithContext(ctx).Where(queryWrapper.StrategyGroup.ID.Eq(params.ID)).Delete(); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (s strategyGroupRepositoryImpl) GetStrategyGroup(ctx context.Context, params *bo.GetStrategyGroupDetailParams) (*bizmodel.StrategyGroup, error) {
	bizDB, err := s.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	bizWrapper := bizQuery.StrategyGroup.WithContext(ctx)
	return bizWrapper.Where(bizQuery.StrategyGroup.ID.Eq(params.ID)).Preload(field.Associations).First()
}

func (s strategyGroupRepositoryImpl) StrategyGroupPage(ctx context.Context, params *bo.QueryStrategyGroupListParams) ([]*bizmodel.StrategyGroup, error) {
	bizDB, err := s.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
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

func (s strategyGroupRepositoryImpl) UpdateStatus(ctx context.Context, params *bo.UpdateStrategyGroupStatusParams) error {
	bizDB, err := s.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
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
