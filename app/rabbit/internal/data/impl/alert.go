package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/convert"
	"github.com/aide-family/rabbit/internal/data/impl/do"
	"github.com/aide-family/rabbit/internal/data/impl/query"
)

func NewAlertSubscriptionRepository(d *data.Data) repository.AlertSubscription {
	query.SetDefault(d.DB())
	return &alertSubscriptionRepository{Data: d}
}

type alertSubscriptionRepository struct {
	*data.Data
}

func (r *alertSubscriptionRepository) GetAlertSubscriptionByName(ctx context.Context, name string) (*bo.AlertSubscriptionItemBo, error) {
	ns := contextx.GetNamespace(ctx)
	model, err := query.AlertSubscription.WithContext(ctx).Where(
		query.AlertSubscription.NamespaceUID.Eq(ns.Int64()),
		query.AlertSubscription.Name.Eq(name),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert subscription not found")
		}
		return nil, err
	}
	return convert.ToAlertSubscriptionItemBo(model), nil
}

func (r *alertSubscriptionRepository) CreateAlertSubscription(ctx context.Context, req *bo.CreateAlertSubscriptionBo) (snowflake.ID, error) {
	model := &do.AlertSubscription{
		NamespaceUID:            contextx.GetNamespace(ctx),
		Name:                    req.Name,
		Remark:                  req.Remark,
		Labels:                  safety.NewMap(req.Labels),
		ExcludeLabels:           safety.NewMap(req.ExcludeLabels),
		RecipientGroupUIDs:      safety.NewSlice(req.RecipientGroupUIDs),
		Members:                 convert.ToNotificationMembersDO(req.Members),
		DirectMemberEmailConfig: req.DirectMemberEmailConfigUID,
		DirectMemberTemplateUID: req.DirectMemberTemplateUID,
		Status:                  enum.GlobalStatus_ENABLED,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	if model.Creator == 0 {
		model.WithCreator(model.NamespaceUID)
	}
	if err := query.AlertSubscription.WithContext(ctx).Create(model); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *alertSubscriptionRepository) UpdateAlertSubscription(ctx context.Context, req *bo.UpdateAlertSubscriptionBo) error {
	ns := contextx.GetNamespace(ctx)
	wrappers := []gen.Condition{
		query.AlertSubscription.NamespaceUID.Eq(ns.Int64()),
		query.AlertSubscription.ID.Eq(req.UID.Int64()),
	}
	columns := []field.AssignExpr{
		query.AlertSubscription.Name.Value(req.Name),
		query.AlertSubscription.Remark.Value(req.Remark),
		query.AlertSubscription.Labels.Value(safety.NewMap(req.Labels)),
		query.AlertSubscription.ExcludeLabels.Value(safety.NewMap(req.ExcludeLabels)),
		query.AlertSubscription.RecipientGroupUIDs.Value(safety.NewSlice(req.RecipientGroupUIDs)),
		query.AlertSubscription.Members.Value(convert.ToNotificationMembersDO(req.Members)),
		query.AlertSubscription.DirectMemberEmailConfig.Value(req.DirectMemberEmailConfigUID.Int64()),
		query.AlertSubscription.DirectMemberTemplateUID.Value(req.DirectMemberTemplateUID.Int64()),
	}
	result, err := query.AlertSubscription.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(columns...)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("alert subscription not found")
	}
	return nil
}

func (r *alertSubscriptionRepository) DeleteAlertSubscription(ctx context.Context, uid snowflake.ID) error {
	ns := contextx.GetNamespace(ctx)
	result, err := query.AlertSubscription.WithContext(ctx).Where(
		query.AlertSubscription.NamespaceUID.Eq(ns.Int64()),
		query.AlertSubscription.ID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("alert subscription not found")
	}
	return nil
}

func (r *alertSubscriptionRepository) GetAlertSubscription(ctx context.Context, uid snowflake.ID) (*bo.AlertSubscriptionDetailBo, error) {
	ns := contextx.GetNamespace(ctx)
	model, err := query.AlertSubscription.WithContext(ctx).Where(
		query.AlertSubscription.NamespaceUID.Eq(ns.Int64()),
		query.AlertSubscription.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("alert subscription not found")
		}
		return nil, err
	}
	recipientGroups, err := r.loadRecipientGroups(ctx, convert.ToAlertSubscriptionRecipientGroupUIDs(model))
	if err != nil {
		return nil, err
	}
	emailConfig, err := r.loadDirectMemberEmailConfig(ctx, model.DirectMemberEmailConfig)
	if err != nil {
		return nil, err
	}
	template, err := r.loadDirectMemberTemplate(ctx, model.DirectMemberTemplateUID)
	if err != nil {
		return nil, err
	}
	return convert.ToAlertSubscriptionDetailBo(model, recipientGroups, emailConfig, template), nil
}

func (r *alertSubscriptionRepository) ListAlertSubscription(ctx context.Context, req *bo.ListAlertSubscriptionBo) (*bo.PageResponseBo[*bo.AlertSubscriptionItemBo], error) {
	ns := contextx.GetNamespace(ctx)
	q := query.AlertSubscription.WithContext(ctx).Where(query.AlertSubscription.NamespaceUID.Eq(ns.Int64()))
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		q = q.Where(
			q.Where(query.AlertSubscription.Name.Like(keyword)).
				Or(query.AlertSubscription.Remark.Like(keyword)),
		)
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		q = q.Where(query.AlertSubscription.Status.Eq(int32(req.Status)))
	}
	total, err := q.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	models, err := q.Order(query.AlertSubscription.CreatedAt.Desc()).Offset(req.Offset()).Limit(req.Limit()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.AlertSubscriptionItemBo, 0, len(models))
	for _, model := range models {
		items = append(items, convert.ToAlertSubscriptionItemBo(model))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *alertSubscriptionRepository) ListEnabledAlertSubscriptions(ctx context.Context) ([]*bo.AlertSubscriptionItemBo, error) {
	ns := contextx.GetNamespace(ctx)
	models, err := query.AlertSubscription.WithContext(ctx).Where(
		query.AlertSubscription.NamespaceUID.Eq(ns.Int64()),
		query.AlertSubscription.Status.Eq(int32(enum.GlobalStatus_ENABLED)),
	).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.AlertSubscriptionItemBo, 0, len(models))
	for _, model := range models {
		items = append(items, convert.ToAlertSubscriptionItemBo(model))
	}
	return items, nil
}

func (r *alertSubscriptionRepository) UpdateAlertSubscriptionStatus(ctx context.Context, req *bo.UpdateAlertSubscriptionStatusBo) error {
	ns := contextx.GetNamespace(ctx)
	result, err := query.AlertSubscription.WithContext(ctx).Where(
		query.AlertSubscription.NamespaceUID.Eq(ns.Int64()),
		query.AlertSubscription.ID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(query.AlertSubscription.Status.Value(int32(req.Status)))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return merr.ErrorNotFound("alert subscription not found")
	}
	return nil
}

func (r *alertSubscriptionRepository) loadRecipientGroups(ctx context.Context, uids []int64) ([]*do.RecipientGroup, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	ids := parseSnowflakeIDs(uids)
	return query.RecipientGroup.WithContext(ctx).Where(
		query.RecipientGroup.NamespaceUID.Eq(ns.Int64()),
		query.RecipientGroup.ID.In(ids...),
	).Preload(field.Associations).Find()
}

func (r *alertSubscriptionRepository) loadDirectMemberEmailConfig(ctx context.Context, uid snowflake.ID) (*do.EmailConfig, error) {
	if uid <= 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	emailConfig, err := query.EmailConfig.WithContext(ctx).Where(
		query.EmailConfig.NamespaceUID.Eq(ns.Int64()),
		query.EmailConfig.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return emailConfig, nil
}

func (r *alertSubscriptionRepository) loadDirectMemberTemplate(ctx context.Context, uid snowflake.ID) (*do.Template, error) {
	if uid <= 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	template, err := query.Template.WithContext(ctx).Where(
		query.Template.NamespaceUID.Eq(ns.Int64()),
		query.Template.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return template, nil
}
