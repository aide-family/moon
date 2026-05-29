package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToRecipientGroupDo(ctx context.Context, req *bo.CreateRecipientGroupBo) *do.RecipientGroup {
	m := &do.RecipientGroup{
		NamespaceUID: contextx.GetNamespace(ctx),
		Name:         req.Name,
		Metadata:     safety.NewMap(req.Metadata),
		Members:      ToNotificationMembersDO(req.Members),
		Status:       enum.GlobalStatus_ENABLED,
	}
	m.WithCreator(contextx.GetUserUID(ctx))
	return m
}

func ToRecipientGroupItemBo(g *do.RecipientGroup) *bo.RecipientGroupItemBo {
	var metadata map[string]string
	if g.Metadata != nil {
		metadata = g.Metadata.Map()
	}
	templates := make([]*bo.TemplateItemBo, 0, len(g.Templates))
	for _, item := range g.Templates {
		if item != nil {
			templates = append(templates, ToTemplateItemBo(item))
		}
	}
	emailConfigs := make([]*bo.EmailConfigItemBo, 0, len(g.EmailConfigs))
	for _, item := range g.EmailConfigs {
		if item != nil {
			emailConfigs = append(emailConfigs, ToEmailConfigBO(item))
		}
	}
	webhookConfigs := make([]*bo.WebhookItemBo, 0, len(g.Webhooks))
	for _, item := range g.Webhooks {
		if item != nil {
			webhookConfigs = append(webhookConfigs, ToWebhookConfigItemBo(item))
		}
	}
	return &bo.RecipientGroupItemBo{
		UID:            g.ID,
		Name:           g.Name,
		Metadata:       metadata,
		Status:         g.Status,
		Templates:      templates,
		EmailConfigs:   emailConfigs,
		WebhookConfigs: webhookConfigs,
		Members:        ToNotificationMembersBo(g.Members),
		CreatedAt:      g.CreatedAt,
		UpdatedAt:      g.UpdatedAt,
	}
}

func toSnowflakeIDsFromTemplates(list []*do.Template) []snowflake.ID {
	var out []snowflake.ID
	for _, t := range list {
		if t != nil {
			out = append(out, t.ID)
		}
	}
	return out
}

func toSnowflakeIDsFromEmailConfigs(list []*do.EmailConfig) []snowflake.ID {
	var out []snowflake.ID
	for _, e := range list {
		if e != nil {
			out = append(out, e.ID)
		}
	}
	return out
}

func toSnowflakeIDsFromWebhookConfigs(list []*do.WebhookConfig) []snowflake.ID {
	var out []snowflake.ID
	for _, w := range list {
		if w != nil {
			out = append(out, w.ID)
		}
	}
	return out
}

// ToRecipientGroupDetailBo converts DO with associations to detail BO.
func ToRecipientGroupDetailBo(g *do.RecipientGroup) *bo.RecipientGroupDetailBo {
	return &bo.RecipientGroupDetailBo{
		RecipientGroupItemBo: *ToRecipientGroupItemBo(g),
		Templates:            toSnowflakeIDsFromTemplates(g.Templates),
		EmailConfigs:         toSnowflakeIDsFromEmailConfigs(g.EmailConfigs),
		WebhookConfigs:       toSnowflakeIDsFromWebhookConfigs(g.Webhooks),
	}
}

func ToRecipientGroupSelectItemBo(g *do.RecipientGroup) *bo.SelectRecipientGroupItemBo {
	return &bo.SelectRecipientGroupItemBo{
		UID:      g.ID,
		Name:     g.Name,
		Status:   g.Status,
		Disabled: g.Status != enum.GlobalStatus_ENABLED || g.DeletedAt.Valid,
		Tooltip:  g.Name,
	}
}

// ToRecipientGroupMetadata converts metadata for update.
func ToRecipientGroupMetadata(m map[string]string) *safety.Map[string, string] {
	return safety.NewMap(m)
}
