package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

func NewTemplateRepository(data *data.Data) repository.Template {
	return &templateRepositoryImpl{data: data}
}

type templateRepositoryImpl struct {
	data *data.Data
}

func (l *templateRepositoryImpl) CreateTemplateStrategy(ctx context.Context, createParam *bo.CreateTemplateStrategyParams) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		templateStrategy := createTemplateStrategy(createParam)
		templateStrategy.WithContext(ctx)
		// 创建主数据
		if err := tx.StrategyTemplate.WithContext(ctx).Create(templateStrategy); err != nil {
			return err
		}
		StrategyTemplateID := templateStrategy.ID
		strategyLevelTemplates := createTemplateLevelParamsToModel(ctx, createParam.StrategyLevelTemplates, StrategyTemplateID)
		// 创建子数据
		if err := tx.StrategyLevelTemplate.WithContext(ctx).Create(strategyLevelTemplates...); err != nil {
			return err
		}
		// 创建关联策略类型
		strategyTemplateCategories := createTemplateCategoriesToModel(ctx, createParam.CategoriesIDs, StrategyTemplateID)

		if err := tx.StrategyTemplateCategories.WithContext(ctx).Create(strategyTemplateCategories...); err != nil {
			return err
		}
		return nil
	})
}

func (l *templateRepositoryImpl) UpdateTemplateStrategy(ctx context.Context, updateParam *bo.UpdateTemplateStrategyParams) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 删除全部关联模板等级数据
		if _, err := tx.StrategyLevelTemplate.WithContext(ctx).Where(query.StrategyLevelTemplate.StrategyTemplateID.Eq(updateParam.ID)).Delete(); err != nil {
			return err
		}
		strategyLevelTemplates := createTemplateLevelParamsToModel(ctx, updateParam.Data.StrategyLevelTemplates, updateParam.ID)
		if err := tx.StrategyLevelTemplate.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(strategyLevelTemplates...); err != nil {
			return err
		}
		_, err := tx.StrategyTemplate.WithContext(ctx).
			Where(query.StrategyTemplate.ID.Eq(updateParam.ID)).
			UpdateSimple(
				query.StrategyTemplate.Expr.Value(updateParam.Data.Expr),
				query.StrategyTemplate.Remark.Value(updateParam.Data.Remark),
				query.StrategyTemplate.Labels.Value(updateParam.Data.Labels),
				query.StrategyTemplate.Annotations.Value(updateParam.Data.Annotations),
				query.StrategyTemplate.Alert.Value(updateParam.Data.Alert),
				query.StrategyTemplate.Status.Value(updateParam.Data.Status.GetValue()),
			)
		// 删除全部关联模板类型数据
		if _, err = tx.StrategyTemplateCategories.WithContext(ctx).Where(query.StrategyTemplateCategories.StrategyTemplateID.Eq(updateParam.ID)).Delete(); err != nil {
			return err
		}

		strategyTemplateCategories := createTemplateCategoriesToModel(ctx, updateParam.Data.CategoriesIDs, updateParam.ID)

		if err := tx.StrategyTemplateCategories.WithContext(ctx).Create(strategyTemplateCategories...); err != nil {
			return err
		}
		return err
	})
}

func (l *templateRepositoryImpl) DeleteTemplateStrategy(ctx context.Context, id uint32) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 删除模板等级关联数据
		if _, err := tx.StrategyLevelTemplate.WithContext(ctx).Where(query.StrategyLevelTemplate.StrategyTemplateID.Eq(id)).Delete(); err != nil {
			return err
		}

		// 删除模板类型关联数据
		if _, err := tx.StrategyTemplateCategories.WithContext(ctx).Where(query.StrategyTemplateCategories.StrategyTemplateID.Eq(id)).Delete(); err != nil {
			return err
		}

		// 删除策略
		_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).
			StrategyTemplate.
			Where(query.StrategyTemplate.ID.Eq(id)).
			Delete()
		return err
	})
}

func (l *templateRepositoryImpl) GetTemplateStrategy(ctx context.Context, id uint32) (*model.StrategyTemplate, error) {
	q := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).StrategyTemplate
	return q.Preload(field.Associations).Preload(query.StrategyTemplate.StrategyLevelTemplates.Level).
		Where(query.StrategyTemplate.ID.Eq(id)).
		First()
}

func (l *templateRepositoryImpl) ListTemplateStrategy(ctx context.Context, params *bo.QueryTemplateStrategyListParams) ([]*model.StrategyTemplate, error) {
	strategyWrapper := query.Use(l.data.GetMainDB(ctx)).StrategyTemplate.WithContext(ctx)

	var wheres []gen.Condition
	if !types.TextIsNull(params.Alert) {
		wheres = append(wheres, query.StrategyTemplate.Alert.Like(params.Alert))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, query.StrategyTemplate.Status.Eq(params.Status.GetValue()))
	}

	if !types.TextIsNull(params.Keyword) {
		strategyWrapper = strategyWrapper.Or(query.StrategyTemplate.Alert.Like(params.Keyword))
		strategyWrapper = strategyWrapper.Or(query.StrategyTemplate.Expr.Like(params.Keyword))
		strategyWrapper = strategyWrapper.Or(query.StrategyTemplate.Remark.Like(params.Keyword))

		dictWrapper := query.Use(l.data.GetMainDB(ctx)).SysDict.WithContext(ctx)

		dictWrapper = dictWrapper.Or(query.SysDict.Name.Like(params.Keyword))
		dictWrapper = dictWrapper.Or(query.SysDict.Value.Like(params.Keyword))
		dictWrapper = dictWrapper.Or(query.SysDict.Remark.Like(params.Keyword))

		sysDicts, err := dictWrapper.Find()
		if err != nil {
			return nil, err
		}

		categoriesIds := types.SliceTo(sysDicts, func(item *model.SysDict) uint32 {
			return item.ID
		})

		strategyWrapper = strategyWrapper.Join(query.StrategyTemplateCategories, query.StrategyTemplateCategories.CategoriesID.In(categoriesIds...))

	}

	strategyWrapper = strategyWrapper.Where(wheres...).Preload(query.StrategyTemplate.StrategyLevelTemplates.Level)

	if err := types.WithPageQuery[query.IStrategyTemplateDo](strategyWrapper, params.Page); err != nil {
		return nil, err
	}
	return strategyWrapper.Order(query.StrategyTemplate.ID).Find()
}

func (l *templateRepositoryImpl) UpdateTemplateStrategyStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).
		StrategyTemplate.
		Where(query.StrategyTemplate.ID.In(ids...)).
		UpdateSimple(query.StrategyTemplate.Status.Value(status.GetValue()))
	return err
}

func createTemplateStrategy(createParam *bo.CreateTemplateStrategyParams) *model.StrategyTemplate {
	return &model.StrategyTemplate{
		Alert:       createParam.Alert,
		Expr:        createParam.Expr,
		Status:      createParam.Status,
		Remark:      createParam.Remark,
		Labels:      createParam.Labels,
		Annotations: createParam.Annotations,
	}
}

func createTemplateLevelParamsToModel(ctx context.Context, params []*bo.CreateStrategyLevelTemplate, templateID uint32) []*model.StrategyLevelTemplate {
	strategyLevelTemplates := types.SliceTo(params, func(item *bo.CreateStrategyLevelTemplate) *model.StrategyLevelTemplate {
		templateLevel := &model.StrategyLevelTemplate{
			StrategyTemplateID: templateID,
			Duration:           item.Duration,
			Count:              item.Count,
			SustainType:        item.SustainType,
			Interval:           item.Interval,
			Condition:          item.Condition,
			Threshold:          item.Threshold,
			LevelID:            item.LevelID,
			Status:             item.Status,
		}
		templateLevel.WithContext(ctx)
		return templateLevel
	})
	return strategyLevelTemplates
}

func createTemplateCategoriesToModel(ctx context.Context, categoriesIds []uint32, templateID uint32) []*model.StrategyTemplateCategories {
	return types.SliceTo(categoriesIds, func(id uint32) *model.StrategyTemplateCategories {
		templateCategories := &model.StrategyTemplateCategories{
			CategoriesID:       id,
			StrategyTemplateID: templateID,
		}
		templateCategories.WithContext(ctx)
		return templateCategories
	})
}
