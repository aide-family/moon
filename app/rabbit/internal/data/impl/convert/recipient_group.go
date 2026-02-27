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
	return &bo.RecipientGroupItemBo{
		UID:      g.ID,
		Name:     g.Name,
		Metadata: metadata,
		Status:   g.Status,
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

func toSnowflakeIDsFromRecipientMembers(list []*do.RecipientMember) []snowflake.ID {
	var out []snowflake.ID
	for _, m := range list {
		if m != nil {
			out = append(out, m.ID)
		}
	}
	return out
}

// ToRecipientGroupDetailBo 从 do（含关联）转为详情 BO
func ToRecipientGroupDetailBo(g *do.RecipientGroup) *bo.RecipientGroupDetailBo {
	return &bo.RecipientGroupDetailBo{
		RecipientGroupItemBo: *ToRecipientGroupItemBo(g),
		Templates:            toSnowflakeIDsFromTemplates(g.Templates),
		EmailConfigs:         toSnowflakeIDsFromEmailConfigs(g.EmailConfigs),
		WebhookConfigs:       toSnowflakeIDsFromWebhookConfigs(g.Webhooks),
		Members:              toSnowflakeIDsFromRecipientMembers(g.Members),
	}
}

// ToRecipientGroupDetailBoFromDo 从 do + 已加载关联转为详情 BO
func ToRecipientGroupDetailBoFromDo(g *do.RecipientGroup, templates []*do.Template, emailConfigs []*do.EmailConfig, webhooks []*do.WebhookConfig, members []*do.RecipientMember) *bo.RecipientGroupDetailBo {
	return &bo.RecipientGroupDetailBo{
		RecipientGroupItemBo: *ToRecipientGroupItemBo(g),
		Templates:            toSnowflakeIDsFromTemplates(templates),
		EmailConfigs:         toSnowflakeIDsFromEmailConfigs(emailConfigs),
		WebhookConfigs:       toSnowflakeIDsFromWebhookConfigs(webhooks),
		Members:              toSnowflakeIDsFromRecipientMembers(members),
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

// ToRecipientGroupMetadata 用于 Update 的 metadata 字段
func ToRecipientGroupMetadata(m map[string]string) *safety.Map[string, string] {
	if m == nil {
		return nil
	}
	return safety.NewMap(m)
}
