package notification

import (
	"context"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
	pkgnotify "github.com/aide-family/moon/pkg/notify"
)

const (
	maxGroupMembers  = 100
	maxGroupWebhooks = 20
	maxGroupChannels = 32
)

// normalizeGroup 校验渠道、模板、成员和 Webhook。
// channel_uids 与 template_uids 一一对应：邮件渠道、短信渠道和每个 Webhook 接收者各一项；
// 邮件和短信同类型最多一个，Webhook 可多选。成员只取当前 Namespace 的成员。
func (center *Center) normalizeGroup(ctx context.Context, in *bo.NotificationGroup) error {
	if strings.TrimSpace(in.Name) == "" {
		return merr.Invalid("name is required")
	}
	in.Name = strings.TrimSpace(in.Name)
	// 渠道类型和对应模板类型必须在套餐范围内：未授权渠道的模板不能绑定到通知组。
	channelTypes, err := center.authorizedChannelTypes(ctx)
	if err != nil {
		return err
	}
	allowedChannels := make(map[string]struct{}, len(channelTypes))
	for _, item := range channelTypes {
		allowedChannels[item] = struct{}{}
	}
	allowedTemplates := make(map[string]struct{})
	for _, item := range perm.TemplateTypesOfChannelTypes(channelTypes) {
		allowedTemplates[item] = struct{}{}
	}
	normalizeLegacyGroup(in)
	// channel_uids 与 template_uids 按下标对应，去重时必须成对删除。
	dedupedChannels := make([]string, 0, len(in.ChannelUIDs))
	dedupedTemplates := make([]string, 0, len(in.TemplateUIDs))
	seenChannels := map[string]struct{}{}
	for index, uid := range in.ChannelUIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, ok := seenChannels[uid]; ok {
			continue
		}
		seenChannels[uid] = struct{}{}
		dedupedChannels = append(dedupedChannels, uid)
		if index < len(in.TemplateUIDs) {
			dedupedTemplates = append(dedupedTemplates, strings.TrimSpace(in.TemplateUIDs[index]))
		} else {
			dedupedTemplates = append(dedupedTemplates, "")
		}
	}
	in.ChannelUIDs = dedupedChannels
	in.TemplateUIDs = dedupedTemplates
	in.EmailMemberUIDs = uniqueStrings(in.EmailMemberUIDs)
	in.SMSMemberUIDs = uniqueStrings(in.SMSMemberUIDs)
	if len(in.ChannelUIDs) == 0 {
		return merr.Invalid("notification group requires at least one channel")
	}
	if len(in.ChannelUIDs) > maxGroupChannels {
		return merr.Invalid("too many channels")
	}
	if len(in.TemplateUIDs) != len(in.ChannelUIDs) {
		return merr.Invalid("channel and template counts do not match")
	}
	for _, uid := range in.TemplateUIDs {
		if strings.TrimSpace(uid) == "" {
			return merr.Invalid("template is required for every channel")
		}
	}
	if len(in.EmailMemberUIDs) > maxGroupMembers || len(in.SMSMemberUIDs) > maxGroupMembers {
		return merr.Invalid("too many members")
	}

	members, err := center.groupMembers(ctx, in)
	if err != nil {
		return err
	}

	emailChannelUID := ""
	smsChannelUID := ""
	webhookUIDs := make([]string, 0, len(in.ChannelUIDs))
	webhookTemplates := map[string]struct{}{}
	for index, channelUID := range in.ChannelUIDs {
		channel, err := center.repo.GetChannel(ctx, in.NamespaceUID, channelUID)
		if err != nil {
			return center.access.Wrap(err)
		}
		template, err := center.repo.GetTemplate(ctx, in.NamespaceUID, in.TemplateUIDs[index])
		if err != nil {
			return center.access.Wrap(err)
		}
		if _, ok := allowedChannels[channel.Type]; !ok {
			return merr.Forbidden("feature not enabled")
		}
		if _, ok := allowedTemplates[template.Type]; !ok {
			return merr.Forbidden("feature not enabled")
		}
		if err := validateGroupBinding(*channel, *template); err != nil {
			return err
		}
		switch channel.Type {
		case channelTypeSMTP:
			if emailChannelUID != "" {
				return merr.Invalid("only one email channel is allowed")
			}
			emailChannelUID = channelUID
		case channelTypeSMS:
			if smsChannelUID != "" {
				return merr.Invalid("only one sms channel is allowed")
			}
			smsChannelUID = channelUID
		case channelTypeWebhook:
			webhookUIDs = append(webhookUIDs, channelUID)
			webhookTemplates[template.UID] = struct{}{}
		}
	}
	if len(webhookUIDs) > maxGroupWebhooks {
		return merr.Invalid("too many webhook channels")
	}

	emailReceivers := mergeReceivers(in.EmailReceivers, nil)
	smsReceivers := mergeReceivers(in.SMSReceivers, nil)
	if emailChannelUID == "" {
		if emailReceivers != "" || len(in.EmailMemberUIDs) > 0 {
			return merr.Invalid("email receivers require an email channel")
		}
	} else {
		emails, err := center.memberEmails(members, in.EmailMemberUIDs)
		if err != nil {
			return err
		}
		merged := mergeReceivers(emailReceivers, emails)
		if merged == "" {
			return merr.Invalid("email channel requires at least one receiver")
		}
		if err := validateReceivers(channelTypeSMTP, merged); err != nil {
			return err
		}
	}
	if smsChannelUID == "" {
		if smsReceivers != "" || len(in.SMSMemberUIDs) > 0 {
			return merr.Invalid("sms receivers require an sms channel")
		}
	} else {
		phones, err := center.memberPhones(members, in.SMSMemberUIDs)
		if err != nil {
			return err
		}
		merged := mergeReceivers(smsReceivers, phones)
		if merged == "" {
			return merr.Invalid("sms channel requires at least one receiver")
		}
		if err := validateReceivers(channelTypeSMS, merged); err != nil {
			return err
		}
	}
	// 只保存手填接收方；成员邮箱/手机号在投递时按 member_uids 实时解析，
	// 避免编辑时把成员地址复制进手填框。
	in.EmailReceivers = emailReceivers
	in.SMSReceivers = smsReceivers

	// 兼容旧客户端和旧版投递字段：写入第一组渠道/模板和 Webhook 快照。
	in.ChannelUID = in.ChannelUIDs[0]
	in.TemplateUID = in.TemplateUIDs[0]
	in.Receivers = in.EmailReceivers
	in.MemberUIDs = in.EmailMemberUIDs
	in.WebhookChannelUIDs = webhookUIDs
	in.WebhookTemplateUID = ""
	if len(webhookTemplates) == 1 {
		for uid := range webhookTemplates {
			in.WebhookTemplateUID = uid
		}
	}
	return nil
}

// normalizeLegacyGroup 把旧版单渠道字段并入新数组，方便旧客户端继续保存。
func normalizeLegacyGroup(in *bo.NotificationGroup) {
	if len(in.ChannelUIDs) > 0 {
		return
	}
	if strings.TrimSpace(in.ChannelUID) == "" {
		return
	}
	in.ChannelUIDs = []string{in.ChannelUID}
	in.TemplateUIDs = []string{in.TemplateUID}
	in.EmailReceivers = in.Receivers
	in.EmailMemberUIDs = in.MemberUIDs
	for _, uid := range in.WebhookChannelUIDs {
		if strings.TrimSpace(uid) == "" {
			continue
		}
		in.ChannelUIDs = append(in.ChannelUIDs, uid)
		in.TemplateUIDs = append(in.TemplateUIDs, in.WebhookTemplateUID)
	}
}

// groupMembers 一次读取邮件和短信选择的成员，避免逐个查询。
func (center *Center) groupMembers(ctx context.Context, in *bo.NotificationGroup) (map[string]bo.Member, error) {
	uids := uniqueStrings(append(append([]string{}, in.EmailMemberUIDs...), in.SMSMemberUIDs...))
	if len(uids) == 0 {
		return map[string]bo.Member{}, nil
	}
	members, err := center.members.ListMembersByUIDs(ctx, in.NamespaceUID, uids)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if len(members) != len(uids) {
		return nil, merr.Invalid("member not found")
	}
	out := make(map[string]bo.Member, len(members))
	for _, item := range members {
		out[item.UID] = item
	}
	return out, nil
}

func (center *Center) memberEmails(members map[string]bo.Member, uids []string) ([]string, error) {
	out := make([]string, 0, len(uids))
	for _, uid := range uids {
		item, ok := members[uid]
		if !ok {
			return nil, merr.Invalid("member not found")
		}
		if item.Status != "JOINED" {
			return nil, merr.Invalid("member is not joined")
		}
		if strings.TrimSpace(item.Email) == "" {
			return nil, merr.Invalid("member has no email")
		}
		out = append(out, item.Email)
	}
	return out, nil
}

func (center *Center) memberPhones(members map[string]bo.Member, uids []string) ([]string, error) {
	out := make([]string, 0, len(uids))
	for _, uid := range uids {
		item, ok := members[uid]
		if !ok {
			return nil, merr.Invalid("member not found")
		}
		if item.Status != "JOINED" {
			return nil, merr.Invalid("member is not joined")
		}
		if strings.TrimSpace(item.Phone) == "" {
			return nil, merr.Invalid("member has no phone")
		}
		if err := validatePhone(item.Phone); err != nil {
			return nil, err
		}
		out = append(out, strings.TrimSpace(item.Phone))
	}
	return out, nil
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
