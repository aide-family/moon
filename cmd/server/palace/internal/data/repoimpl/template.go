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

func (l *templateRepositoryImpl) CreateTemplateStrategy(ctx context.Context, templateStrategy *bo.CreateTemplateStrategyParams) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建主数据
		if err := tx.StrategyTemplate.WithContext(ctx).Create(templateStrategy.StrategyTemplate); err != nil {
			return err
		}

		strategyLevelTemplates := types.SliceTo(templateStrategy.StrategyLevelTemplates, func(item *model.StrategyLevelTemplate) *model.StrategyLevelTemplate {
			item.StrategyID = templateStrategy.ID
			return item
		})
		// 创建子数据
		if err := tx.StrategyLevelTemplate.WithContext(ctx).Create(strategyLevelTemplates...); err != nil {
			return err
		}
		return nil
	})
}

func (l *templateRepositoryImpl) UpdateTemplateStrategy(ctx context.Context, templateStrategy *bo.UpdateTemplateStrategyParams) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 删除全部关联数据
		if _, err := tx.StrategyLevelTemplate.WithContext(ctx).Where(query.StrategyLevelTemplate.StrategyID.Eq(templateStrategy.ID)).Delete(); err != nil {
			return err
		}
		if err := tx.StrategyLevelTemplate.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(templateStrategy.StrategyLevelTemplates...); err != nil {
			return err
		}
		_, err := tx.StrategyTemplate.WithContext(ctx).
			Where(query.StrategyTemplate.ID.Eq(templateStrategy.ID)).
			UpdateSimple(
				query.StrategyTemplate.Expr.Value(templateStrategy.Expr),
				query.StrategyTemplate.Remark.Value(templateStrategy.Remark),
				query.StrategyTemplate.Labels.Value(templateStrategy.Labels),
				query.StrategyTemplate.Annotations.Value(templateStrategy.Annotations),
				query.StrategyTemplate.Alert.Value(templateStrategy.Alert),
				query.StrategyTemplate.Status.Value(templateStrategy.Status.GetValue()),
			)
		return err
	})
}

func (l *templateRepositoryImpl) DeleteTemplateStrategy(ctx context.Context, id uint32) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 删除关联数据
		if _, err := tx.StrategyLevelTemplate.WithContext(ctx).Where(query.StrategyLevelTemplate.StrategyID.Eq(id)).Delete(); err != nil {
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
	return q.Preload(field.Associations).
		Where(query.StrategyTemplate.ID.Eq(id)).
		First()
}

func (l *templateRepositoryImpl) ListTemplateStrategy(ctx context.Context, params *bo.QueryTemplateStrategyListParams) ([]*model.StrategyTemplate, error) {
	q := query.Use(l.data.GetMainDB(ctx)).StrategyTemplate
	qq := q.WithContext(ctx)

	var wheres []gen.Condition
	if !types.TextIsNull(params.Alert) {
		wheres = append(wheres, query.StrategyTemplate.Alert.Like(params.Alert))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, query.StrategyTemplate.Status.Eq(params.Status.GetValue()))
	}
	qq = qq.Where(wheres...)
	if !types.IsNil(params.Page) {
		page := params.Page
		total, err := qq.WithContext(ctx).Count()
		if !types.IsNil(err) {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			qq = qq.Limit(pageSize)
		} else {
			qq = qq.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return qq.Order(q.ID.Desc()).Find()
}

func (l *templateRepositoryImpl) UpdateTemplateStrategyStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).
		StrategyTemplate.
		Where(query.StrategyTemplate.ID.In(ids...)).
		UpdateSimple(query.StrategyTemplate.Status.Value(status.GetValue()))
	return err
}
