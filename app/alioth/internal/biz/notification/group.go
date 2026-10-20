package notification

import (
	"context"
	"strings"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
	pkgnotify "github.com/aide-family/Alioth/pkg/notify"
)

const (
	maxGroupMembers  = 100
	maxGroupWebhooks = 20
)

// normalizeGroup 校验主渠道、模板、成员和 Webhook，并把重复项去掉。
// 成员只取当前 Namespace 的成员，避免跨租户读取用户。
func (uc *Usecase) normalizeGroup(ctx context.Context, in *bo.NotificationGroup) error {
	if strings.TrimSpace(in.Name) == "" {
		return merr.Invalid("name is required")
	}
	in.Name = strings.TrimSpace(in.Name)
	channel, err := uc.repo.GetChannel(ctx, in.NamespaceUID, in.ChannelUID)
	if err != nil {
		return uc.access.Wrap(err)
	}
	template, err := uc.repo.GetTemplate(ctx, in.NamespaceUID, in.TemplateUID)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if err := validateGroupBinding(*channel, *template); err != nil {
		return err
	}

	in.MemberUIDs = uniqueStrings(in.MemberUIDs)
	in.WebhookChannelUIDs = uniqueStrings(in.WebhookChannelUIDs)
	if len(in.MemberUIDs) > maxGroupMembers {
		return merr.Invalid("too many members")
	}
	if len(in.WebhookChannelUIDs) > maxGroupWebhooks {
		return merr.Invalid("too many webhook channels")
	}
	memberEmails := make([]string, 0, len(in.MemberUIDs))
	if len(in.MemberUIDs) > 0 {
		if channel.Type != channelTypeSMTP {
			return merr.Invalid("members require an smtp channel")
		}
		members, err := uc.members.ListMembersByUIDs(ctx, in.NamespaceUID, in.MemberUIDs)
		if err != nil {
			return uc.access.Wrap(err)
		}
		if len(members) != len(in.MemberUIDs) {
			return merr.Invalid("member not found")
		}
		for _, item := range members {
			if item.Status != "JOINED" {
				return merr.Invalid("member is not joined")
			}
			if strings.TrimSpace(item.Email) == "" {
				return merr.Invalid("member has no email")
			}
			memberEmails = append(memberEmails, item.Email)
		}
	}
	if err := validateReceivers(channel.Type, mergeReceivers(in.Receivers, memberEmails)); err != nil {
		return err
	}

	if channel.Type == channelTypeWebhook {
		for i, uid := range in.WebhookChannelUIDs {
			if uid == in.ChannelUID {
				in.WebhookChannelUIDs = append(in.WebhookChannelUIDs[:i], in.WebhookChannelUIDs[i+1:]...)
				break
			}
		}
	}
	if len(in.WebhookChannelUIDs) == 0 {
		in.WebhookTemplateUID = ""
		return nil
	}
	if strings.TrimSpace(in.WebhookTemplateUID) == "" {
		return merr.Invalid("webhook template is required")
	}
	webhookTemplate, err := uc.repo.GetTemplate(ctx, in.NamespaceUID, in.WebhookTemplateUID)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if webhookTemplate.Status != "ACTIVE" || webhookTemplate.Type != "WEBHOOK" {
		return merr.Invalid("invalid webhook template")
	}
	for _, uid := range in.WebhookChannelUIDs {
		webhook, err := uc.repo.GetChannel(ctx, in.NamespaceUID, uid)
		if err != nil {
			return uc.access.Wrap(err)
		}
		if webhook.Type != channelTypeWebhook || webhook.Status != "ACTIVE" {
			return merr.Invalid("invalid webhook channel")
		}
	}
	return nil
}

func mergeReceivers(extra string, emails []string) string {
	values := append(pkgnotify.SplitReceivers(extra), emails...)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return strings.Join(out, ",")
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
