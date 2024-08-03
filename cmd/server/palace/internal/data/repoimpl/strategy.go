package repoimpl

import (
	"context"
	"fmt"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
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

func (s *strategyRepositoryImpl) UpdateStatus(ctx context.Context, params *bo.UpdateStrategyStatusParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.WithContext(ctx).
		Strategy.Where(bizQuery.Strategy.ID.In(params.Ids...)).
		Update(bizQuery.Strategy.Status, params.Status)
	return err
}

func (s *strategyRepositoryImpl) DeleteByID(ctx context.Context, strategyID uint32) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if _, err = tx.Strategy.WithContext(ctx).Where(tx.Strategy.ID.Eq(strategyID)).Delete(); !types.IsNil(err) {
			return err
		}
		if _, err = tx.StrategyLevel.WithContext(ctx).Where(tx.StrategyLevel.StrategyID.Eq(strategyID)).Delete(); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (s *strategyRepositoryImpl) CreateStrategy(ctx context.Context, params *bo.CreateStrategyParams) (*bizmodel.Strategy, error) {
	bizQuery, err := getBizQuery(ctx, s.data)
	templateID := params.TemplateID
	if !types.IsNil(err) {
		return nil, err
	}
	mainDb := s.data.GetMainDB(ctx).WithContext(ctx)
	mainQuery := query.Use(mainDb)
	strategyTemplate, err := mainQuery.WithContext(ctx).StrategyTemplate.Where(mainQuery.StrategyTemplate.ID.Eq(templateID)).Preload(field.Associations).First()
	if !types.IsNil(err) {
		return nil, err
	}

	strategyModel := createStrategyParamsToModel(ctx, strategyTemplate, params)

	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.Strategy.WithContext(ctx).Create(strategyModel); !types.IsNil(err) {
			return err
		}
		// Creating  Strategy levels
		strategyLevels := createStrategyLevelParamsToModel(ctx, params.StrategyLevel, strategyModel.ID)
		if err := tx.StrategyLevel.WithContext(ctx).Create(strategyLevels...); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	return strategyModel, nil
}

func (s *strategyRepositoryImpl) UpdateByID(ctx context.Context, params *bo.UpdateStrategyParams) error {
	bizQuery, err := getBizQuery(ctx, s.data)
	if !types.IsNil(err) {
		return err
	}
	updateParam := params.UpdateParam
	datasourceIds := types.SliceToWithFilter(updateParam.DatasourceIDs, func(datasourceId uint32) (*bizmodel.Datasource, bool) {
		if datasourceId <= 0 {
			return nil, false
		}
		return &bizmodel.Datasource{
			AllFieldModel: model.AllFieldModel{ID: datasourceId},
		}, true
	})
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err = tx.Strategy.Datasource.
			Model(&bizmodel.Strategy{AllFieldModel: model.AllFieldModel{ID: params.ID}}).Replace(datasourceIds...); !types.IsNil(err) {
			return err
		}

		strategyTemplate, err := tx.StrategyTemplate.Where(tx.StrategyTemplate.ID.Eq(params.UpdateParam.TemplateID)).Preload(field.Associations).First()

		if strategyTemplate != nil {
			categories := types.SliceToWithFilter(strategyTemplate.Categories, func(dict *bizmodel.SysDict) (*bizmodel.SysDict, bool) {
				if dict.ID <= 0 {
					return nil, false
				}
				return &bizmodel.SysDict{
					AllFieldModel: model.AllFieldModel{ID: dict.ID},
				}, true
			})

			if err = tx.Strategy.Categories.Model(&bizmodel.Strategy{AllFieldModel: model.AllFieldModel{ID: params.ID}}).Replace(categories...); !types.IsNil(err) {
				return err
			}
		}
		// 删除策略等级数据
		if _, err = tx.StrategyLevel.WithContext(ctx).Where(tx.StrategyLevel.StrategyID.Eq(params.ID)).Delete(); !types.IsNil(err) {
			return err
		}
		// Creating  Strategy levels
		strategyLevels := createStrategyLevelParamsToModel(ctx, updateParam.StrategyLevel, params.ID)
		if err = tx.StrategyLevel.WithContext(ctx).Create(strategyLevels...); !types.IsNil(err) {
			return err
		}

		// 更新策略
		if _, err = tx.Strategy.WithContext(ctx).Where(tx.Strategy.ID.Eq(params.ID)).UpdateSimple(
			tx.Strategy.Name.Value(updateParam.Name),
			tx.Strategy.Step.Value(updateParam.Step),
			tx.Strategy.Remark.Value(updateParam.Remark),
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
	return bizWrapper.Where(bizQuery.Strategy.ID.Eq(strategyID)).Preload(field.Associations).First()
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
	return strategy, nil
}

func createStrategyLevelParamsToModel(ctx context.Context, params []*bo.CreateStrategyLevel, strategyID uint32) []*bizmodel.StrategyLevel {
	strategyLevel := types.SliceTo(params, func(item *bo.CreateStrategyLevel) *bizmodel.StrategyLevel {
		templateLevel := &bizmodel.StrategyLevel{
			StrategyID:  strategyID,
			Duration:    item.Duration,
			Count:       item.Count,
			SustainType: item.SustainType,
			Interval:    item.Interval,
			Condition:   item.Condition,
			Threshold:   item.Threshold,
			LevelID:     item.LevelID,
			Status:      item.Status,
		}
		templateLevel.WithContext(ctx)
		return templateLevel
	})
	return strategyLevel
}

func createStrategyParamsToModel(ctx context.Context, strategyTemplate *model.StrategyTemplate, params *bo.CreateStrategyParams) *bizmodel.Strategy {
	strategyModel := &bizmodel.Strategy{
		Name:                   params.Name,
		StrategyTemplateID:     strategyTemplate.ID,
		StrategyTemplateSource: vobj.StrategyTemplateSource(params.SourceType),
		Expr:                   strategyTemplate.Expr,
		Labels:                 params.Labels,
		Annotations:            params.Annotations,
		Remark:                 params.Remark,
		Status:                 vobj.Status(params.Status.GetValue()),
		Step:                   params.Step,
		Datasource: types.SliceToWithFilter(params.DatasourceIDs, func(datasourceId uint32) (*bizmodel.Datasource, bool) {
			if datasourceId <= 0 {
				return nil, false
			}
			return &bizmodel.Datasource{
				AllFieldModel: model.AllFieldModel{ID: datasourceId},
			}, true
		}),
		Categories: types.SliceToWithFilter(strategyTemplate.Categories, func(dict *model.SysDict) (*bizmodel.SysDict, bool) {
			if dict.ID <= 0 {
				return nil, false
			}
			return &bizmodel.SysDict{
				AllFieldModel: model.AllFieldModel{ID: dict.ID},
			}, true
		}),
		GroupID: params.GroupID,
	}
	strategyModel.WithContext(ctx)
	return strategyModel
}
