package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewAlertPageRepository(d *data.Data) (repository.AlertPage, error) {
	query.SetDefault(d.DB())
	return &alertPageRepository{db: d.DB()}, nil
}

type alertPageRepository struct {
	db *gorm.DB
}

func (r *alertPageRepository) AlertPageNameTaken(ctx context.Context, name string, excludeUID snowflake.ID) (bool, error) {
	a := query.AlertPage
	wrappers := a.WithContext(ctx).Where(
		a.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		a.Name.Eq(name),
		a.ID.Neq(excludeUID.Int64()),
	)
	n, err := wrappers.Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *alertPageRepository) CreateAlertPage(ctx context.Context, req *bo.CreateAlertPageBo) (snowflake.ID, error) {
	m := convert.ToAlertPageDo(ctx, req)
	if err := query.AlertPage.WithContext(ctx).Create(m); err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *alertPageRepository) UpdateAlertPage(ctx context.Context, req *bo.UpdateAlertPageBo) error {
	model, filterConfig := convert.ToAlertPageDoUpdate(req)
	if model == nil {
		return merr.ErrorInvalidArgument("update request is nil")
	}
	a := query.AlertPage
	columns := []field.AssignExpr{
		a.Name.Value(model.Name),
		a.Color.Value(model.Color),
		a.SortOrder.Value(model.SortOrder),
	}
	if filterConfig != nil {
		columns = append(columns, a.FilterConfig.Value(filterConfig))
	}
	_, err := a.WithContext(ctx).Where(
		a.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		a.ID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(columns...)
	if err != nil {
		return err
	}
	return nil
}

func (r *alertPageRepository) DeleteAlertPage(ctx context.Context, uid snowflake.ID) error {
	a := query.AlertPage
	info, err := a.WithContext(ctx).Where(
		a.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		a.ID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("alert page not found")
	}
	return nil
}

func (r *alertPageRepository) GetAlertPage(ctx context.Context, uid snowflake.ID) (*bo.AlertPageItemBo, error) {
	a := query.AlertPage
	m, err := a.WithContext(ctx).Where(
		a.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		a.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert page not found")
		}
		return nil, err
	}
	return convert.ToAlertPageItemBo(m), nil
}

func (r *alertPageRepository) ListAlertPage(ctx context.Context, req *bo.ListAlertPageBo) (*bo.PageResponseBo[*bo.AlertPageItemBo], error) {
	a := query.AlertPage
	wrappers := a.WithContext(ctx).Where(a.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		wrappers = wrappers.Where(a.Name.Like("%" + req.Keyword + "%"))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Order(a.SortOrder).Offset(req.Offset()).Limit(req.Limit())
	}
	list, err := wrappers.Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.AlertPageItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToAlertPageItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}
