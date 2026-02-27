package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
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
	q := query.RecipientGroup.WithContext(ctx).Where(
		query.RecipientGroup.NamespaceUID.Eq(ns.Int64()),
		query.RecipientGroup.Name.Eq(name),
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
	templates := make([]*do.Template, 0, len(req.Templates))
	for _, id := range req.Templates {
		templates = append(templates, &do.Template{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
	}
	emailConfigs := make([]*do.EmailConfig, 0, len(req.EmailConfigs))
	for _, id := range req.EmailConfigs {
		emailConfigs = append(emailConfigs, &do.EmailConfig{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
	}
	webhooks := make([]*do.WebhookConfig, 0, len(req.WebhookConfigs))
	for _, id := range req.WebhookConfigs {
		webhooks = append(webhooks, &do.WebhookConfig{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
	}
	members := make([]*do.RecipientMember, 0, len(req.Members))
	for _, id := range req.Members {
		members = append(members, &do.RecipientMember{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
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
		if err := mutation.Members.WithContext(ctx).Model(m).Replace(members...); err != nil {
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
	group, err := q.Preload(query.RecipientGroup.Templates).
		Preload(query.RecipientGroup.EmailConfigs).
		Preload(query.RecipientGroup.Webhooks).
		Preload(query.RecipientGroup.Members).
		First()
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
	q := queryMutation.WithContext(ctx).Where(
		queryMutation.NamespaceUID.Eq(ns.Int64()),
		queryMutation.ID.Eq(req.UID.Int64()),
	)
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
	}

	return r.DB().Transaction(func(tx *gorm.DB) error {
		mutation := query.Use(tx).RecipientGroup
		if _, err := mutation.UpdateColumnSimple(columns...); err != nil {
			return err
		}

		templates := make([]*do.Template, 0, len(req.Templates))
		for _, id := range req.Templates {
			templates = append(templates, &do.Template{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
		}
		if err := mutation.Templates.WithContext(ctx).Model(group).Replace(templates...); err != nil {
			return err
		}
		configs := make([]*do.EmailConfig, 0, len(req.EmailConfigs))
		for _, id := range req.EmailConfigs {
			configs = append(configs, &do.EmailConfig{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
		}
		if err := mutation.EmailConfigs.WithContext(ctx).Model(group).Replace(configs...); err != nil {
			return err
		}
		webhooks := make([]*do.WebhookConfig, 0, len(req.WebhookConfigs))
		for _, id := range req.WebhookConfigs {
			webhooks = append(webhooks, &do.WebhookConfig{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
		}
		if err := mutation.Webhooks.WithContext(ctx).Model(group).Replace(webhooks...); err != nil {
			return err
		}
		members := make([]*do.RecipientMember, 0, len(req.Members))
		for _, id := range req.Members {
			members = append(members, &do.RecipientMember{BaseModel: do.BaseModel{ID: snowflake.ParseInt64(id)}})
		}
		if err := mutation.Members.WithContext(ctx).Model(group).Replace(members...); err != nil {
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
	list, err := q.Order(query.RecipientGroup.CreatedAt.Desc()).Find()
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
