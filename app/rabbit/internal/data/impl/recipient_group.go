package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
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

func NewRecipientGroupRepository(d *data.Data) repository.RecipientGroup {
	query.SetDefault(d.DB())
	return &recipientGroupRepository{Data: d}
}

type recipientGroupRepository struct {
	*data.Data
}

// GetRecipientGroupByName implements [repository.RecipientGroup].
func (r *recipientGroupRepository) GetRecipientGroupByName(ctx context.Context, name string) (*bo.RecipientGroupItemBo, error) {
	ns := contextx.GetNamespace(ctx)
	recipientGroup := query.RecipientGroup
	q := recipientGroup.WithContext(ctx).Where(
		recipientGroup.NamespaceUID.Eq(ns.Int64()),
		recipientGroup.Name.Eq(name),
	)

	group, err := q.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("recipient group not found")
		}
		return nil, err
	}
	return convert.ToRecipientGroupItemBo(group), nil
}

func (r *recipientGroupRepository) CreateRecipientGroup(ctx context.Context, req *bo.CreateRecipientGroupBo) (uid snowflake.ID, err error) {
	m := convert.ToRecipientGroupDo(ctx, req)
	templates, err := r.loadTemplatesForReplace(ctx, req.Templates)
	if err != nil {
		return 0, err
	}
	emailConfigs, err := r.loadEmailConfigsForReplace(ctx, req.EmailConfigs)
	if err != nil {
		return 0, err
	}
	webhooks, err := r.loadWebhooksForReplace(ctx, req.WebhookConfigs)
	if err != nil {
		return 0, err
	}
	err = r.DB().Transaction(func(tx *gorm.DB) error {
		mutation := query.Use(tx).RecipientGroup
		if err := mutation.WithContext(ctx).Create(m); err != nil {
			return err
		}
		uid = m.ID
		if err := mutation.Templates.WithContext(ctx).Model(m).Replace(templates...); err != nil {
			return err
		}
		if err := mutation.EmailConfigs.WithContext(ctx).Model(m).Replace(emailConfigs...); err != nil {
			return err
		}
		if err := mutation.Webhooks.WithContext(ctx).Model(m).Replace(webhooks...); err != nil {
			return err
		}
		return nil
	})
	return
}

func (r *recipientGroupRepository) GetRecipientGroup(ctx context.Context, uid snowflake.ID) (*bo.RecipientGroupDetailBo, error) {
	ns := contextx.GetNamespace(ctx)
	q := query.RecipientGroup.WithContext(ctx).Where(
		query.RecipientGroup.NamespaceUID.Eq(ns.Int64()),
		query.RecipientGroup.ID.Eq(uid.Int64()),
	)
	group, err := q.Preload(field.Associations).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("recipient group not found")
		}
		return nil, err
	}
	return convert.ToRecipientGroupDetailBo(group), nil
}

func (r *recipientGroupRepository) UpdateRecipientGroup(ctx context.Context, req *bo.UpdateRecipientGroupBo) error {
	ns := contextx.GetNamespace(ctx)
	queryMutation := query.RecipientGroup
	wrappers := []gen.Condition{
		queryMutation.NamespaceUID.Eq(ns.Int64()),
		queryMutation.ID.Eq(req.UID.Int64()),
	}
	q := queryMutation.WithContext(ctx).Where(wrappers...)
	group, err := q.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorNotFound("recipient group not found")
		}
		return err
	}
	columns := []field.AssignExpr{
		queryMutation.Name.Value(req.Name),
		queryMutation.Metadata.Value(convert.ToRecipientGroupMetadata(req.Metadata)),
		queryMutation.Members.Value(convert.ToNotificationMembersDO(req.Members)),
	}

	return r.DB().Transaction(func(tx *gorm.DB) error {
		mutation := query.Use(tx).RecipientGroup
		if _, err := mutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(columns...); err != nil {
			return err
		}

		templates, err := r.loadTemplatesForReplace(ctx, req.Templates)
		if err != nil {
			return err
		}
		if err := mutation.Templates.WithContext(ctx).Model(group).Replace(templates...); err != nil {
			return err
		}
		configs, err := r.loadEmailConfigsForReplace(ctx, req.EmailConfigs)
		if err != nil {
			return err
		}
		if err := mutation.EmailConfigs.WithContext(ctx).Model(group).Replace(configs...); err != nil {
			return err
		}
		webhooks, err := r.loadWebhooksForReplace(ctx, req.WebhookConfigs)
		if err != nil {
			return err
		}
		if err := mutation.Webhooks.WithContext(ctx).Model(group).Replace(webhooks...); err != nil {
			return err
		}
		return nil
	})
}

func (r *recipientGroupRepository) UpdateRecipientGroupStatus(ctx context.Context, req *bo.UpdateRecipientGroupStatusBo) error {
	ns := contextx.GetNamespace(ctx)
	q := query.RecipientGroup.WithContext(ctx).Where(
		query.RecipientGroup.NamespaceUID.Eq(ns.Int64()),
		query.RecipientGroup.ID.Eq(req.UID.Int64()),
	)
	_, err := q.UpdateColumnSimple(query.RecipientGroup.Status.Value(int32(req.Status)))
	return err
}

func (r *recipientGroupRepository) DeleteRecipientGroup(ctx context.Context, uid snowflake.ID) error {
	ns := contextx.GetNamespace(ctx)
	_, err := query.RecipientGroup.WithContext(ctx).Where(
		query.RecipientGroup.NamespaceUID.Eq(ns.Int64()),
		query.RecipientGroup.ID.Eq(uid.Int64()),
	).Delete()
	return err
}

func (r *recipientGroupRepository) ListRecipientGroup(ctx context.Context, req *bo.ListRecipientGroupBo) (*bo.PageResponseBo[*bo.RecipientGroupItemBo], error) {
	ns := contextx.GetNamespace(ctx)
	q := query.RecipientGroup.WithContext(ctx).Where(query.RecipientGroup.NamespaceUID.Eq(ns.Int64()))
	if strutil.IsNotEmpty(req.Keyword) {
		q = q.Where(query.RecipientGroup.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		q = q.Where(query.RecipientGroup.Status.Eq(int32(req.Status)))
	}
	if pointer.IsNotNil(req.PageRequestBo) {
		total, err := q.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		q = q.Limit(req.Limit()).Offset(req.Offset())
	}
	list, err := q.Order(query.RecipientGroup.CreatedAt.Desc()).Preload(field.Associations).Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.NewPageResponseBo(req.PageRequestBo, []*bo.RecipientGroupItemBo{}), nil
		}
		return nil, err
	}
	items := make([]*bo.RecipientGroupItemBo, 0, len(list))
	for _, g := range list {
		items = append(items, convert.ToRecipientGroupItemBo(g))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *recipientGroupRepository) loadEmailConfigsForReplace(ctx context.Context, configIDs []int64) ([]*do.EmailConfig, error) {
	if len(configIDs) == 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	ids := parseSnowflakeIDs(configIDs)
	list, err := query.EmailConfig.WithContext(ctx).Where(
		query.EmailConfig.NamespaceUID.Eq(ns.Int64()),
		query.EmailConfig.ID.In(ids...),
	).Find()
	if err != nil {
		return nil, err
	}
	if len(list) != len(configIDs) {
		return nil, merr.ErrorInvalidArgument("email config not found")
	}
	return list, nil
}

func (r *recipientGroupRepository) loadTemplatesForReplace(ctx context.Context, templateIDs []int64) ([]*do.Template, error) {
	if len(templateIDs) == 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	ids := parseSnowflakeIDs(templateIDs)
	list, err := query.Template.WithContext(ctx).Where(
		query.Template.NamespaceUID.Eq(ns.Int64()),
		query.Template.ID.In(ids...),
	).Find()
	if err != nil {
		return nil, err
	}
	if len(list) != len(templateIDs) {
		return nil, merr.ErrorInvalidArgument("template not found")
	}
	return list, nil
}

func (r *recipientGroupRepository) loadWebhooksForReplace(ctx context.Context, webhookIDs []int64) ([]*do.WebhookConfig, error) {
	if len(webhookIDs) == 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	ids := parseSnowflakeIDs(webhookIDs)
	list, err := query.WebhookConfig.WithContext(ctx).Where(
		query.WebhookConfig.NamespaceUID.Eq(ns.Int64()),
		query.WebhookConfig.ID.In(ids...),
	).Find()
	if err != nil {
		return nil, err
	}
	if len(list) != len(webhookIDs) {
		return nil, merr.ErrorInvalidArgument("webhook config not found")
	}
	return list, nil
}

func parseSnowflakeIDs(raw []int64) []int64 {
	ids := make([]int64, len(raw))
	for i, id := range raw {
		ids[i] = snowflake.ParseInt64(id).Int64()
	}
	return ids
}

func (r *recipientGroupRepository) SelectRecipientGroup(ctx context.Context, req *bo.SelectRecipientGroupBo) (*bo.SelectRecipientGroupBoResult, error) {
	ns := contextx.GetNamespace(ctx)
	q := query.RecipientGroup.WithContext(ctx).Where(query.RecipientGroup.NamespaceUID.Eq(ns.Int64()))
	if strutil.IsNotEmpty(req.Keyword) {
		q = q.Where(query.RecipientGroup.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		q = q.Where(query.RecipientGroup.Status.Eq(int32(req.Status)))
	}
	total, err := q.Count()
	if err != nil {
		return nil, err
	}
	if req.LastUID > 0 {
		q = q.Where(query.RecipientGroup.ID.Lt(req.LastUID.Int64()))
	}
	q = q.Limit(int(req.Limit)).Order(query.RecipientGroup.ID.Desc())
	list, err := q.Find()
	if err != nil {
		return nil, err
	}
	var lastUID snowflake.ID
	if len(list) > 0 {
		lastUID = list[len(list)-1].ID
	}
	items := make([]*bo.SelectRecipientGroupItemBo, 0, len(list))
	for _, g := range list {
		items = append(items, convert.ToRecipientGroupSelectItemBo(g))
	}
	return &bo.SelectRecipientGroupBoResult{
		Items:   items,
		Total:   total,
		LastUID: lastUID,
	}, nil
}
