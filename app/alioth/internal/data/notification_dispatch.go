package data

import (
	"context"
	"net/url"
	"strings"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	pkgnotify "github.com/aide-family/Alioth/pkg/notify"
)

// notifyEvent 是分发订阅时需要的告警事件字段。
type notifyEvent struct {
	NamespaceUID  string
	EventUID      string
	RuleUID       string
	DatasourceUID string
	LevelUID      string
	Phase         string
	Summary       string
	Description   string
	Labels        string
	Value         string
	ExternalURL   string
	Fingerprint   string
}

type notifyTarget struct {
	ChannelUID  string
	TemplateUID string
	Sender      string
	Receivers   string
	Subject     string
	Body        string
}

// notificationSender 返回可供审计展示的实际发送身份，不包含渠道凭据或连接地址。
func notificationSender(channel *ent.NotificationChannel) string {
	if channel.FromAddress != "" {
		return channel.FromAddress
	}
	if channel.Type == notificationchannel.TypeSMTP && channel.Username != "" {
		return channel.Username
	}
	return channel.Name
}

// notificationReceiver 返回可审计的 Webhook 接收端；去除 URL 用户信息、查询参数和 fragment，避免记录密钥。
func notificationReceiver(channel *ent.NotificationChannel) string {
	if channel.Type != notificationchannel.TypeWEBHOOK {
		return ""
	}
	parsed, err := url.Parse(channel.Endpoint)
	if err != nil {
		return channel.Name
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	return parsed.String()
}

// dispatchEventNotifications 按订阅条件把事件发给订阅下的通知组。
// 调用方必须在写入事件的事务内调用，保证事件和任务同时生效。
func dispatchEventNotifications(ctx context.Context, tx *ent.Tx, event notifyEvent) error {
	subscriptions, err := tx.NotificationSubscription.Query().Where(
		notificationsubscription.NamespaceUIDEQ(event.NamespaceUID),
		notificationsubscription.StatusEQ(notificationsubscription.StatusACTIVE),
		notificationsubscription.DeletedAtIsNil(),
	).Order(ent.Asc(notificationsubscription.FieldUID)).All(ctx)
	if err != nil {
		return err
	}
	matched := make([]*ent.NotificationSubscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if matchSubscription(subscription, event) {
			matched = append(matched, subscription)
		}
	}
	if len(matched) == 0 {
		return nil
	}
	subscriptionUIDs := make([]string, len(matched))
	for i, subscription := range matched {
		subscriptionUIDs[i] = subscription.UID
	}
	links, err := tx.NotificationSubscriptionGroup.Query().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(event.NamespaceUID),
		notificationsubscriptiongroup.SubscriptionUIDIn(subscriptionUIDs...),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return err
	}
	// 同一通知组被多个订阅命中时只保留首个订阅标识，任务仍按渠道去重。
	groupSubscription := make(map[string]string, len(links))
	for _, subscription := range matched {
		for _, link := range links {
			if link.SubscriptionUID != subscription.UID {
				continue
			}
			if _, ok := groupSubscription[link.GroupUID]; !ok {
				groupSubscription[link.GroupUID] = subscription.UID
			}
		}
	}
	if len(groupSubscription) == 0 {
		return nil
	}
	groupUIDs := make([]string, 0, len(groupSubscription))
	for groupUID := range groupSubscription {
		groupUIDs = append(groupUIDs, groupUID)
	}
	groups, err := tx.NotificationGroup.Query().Where(
		notificationgroup.UIDIn(groupUIDs...),
		notificationgroup.NamespaceUIDEQ(event.NamespaceUID),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
		notificationgroup.DeletedAtIsNil(),
	).Order(ent.Asc(notificationgroup.FieldUID)).All(ctx)
	if err != nil {
		return err
	}
	params := paramsJSON(event)
	for _, group := range groups {
		targets, err := resolveGroupTargets(ctx, tx, group)
		if err != nil {
			return err
		}
		for _, target := range targets {
			err := tx.NotificationTask.Create().
				SetUID(id.New()).
				SetNamespaceUID(event.NamespaceUID).
				SetEventUID(event.EventUID).
				SetPhase(notificationtask.Phase(event.Phase)).
				SetStatus(notificationtask.StatusPENDING).
				SetRuleUID(event.RuleUID).
				SetSubscriptionUID(groupSubscription[group.UID]).
				SetGroupUID(group.UID).
				SetChannelUID(target.ChannelUID).
				SetTemplateUID(target.TemplateUID).
				SetSender(target.Sender).
				SetReceivers(target.Receivers).
				SetSubject(target.Subject).
				SetBody(target.Body).
				SetContent(target.Body).
				SetParams(params).
				SetRetriesLeft(3).
				Exec(ctx)
			if err != nil {
				if ent.IsConstraintError(err) {
					continue
				}
				return err
			}
		}
	}
	return nil
}

func matchSubscription(subscription *ent.NotificationSubscription, event notifyEvent) bool {
	if len(subscription.LevelUids) > 0 && !containsString(subscription.LevelUids, event.LevelUID) {
		return false
	}
	if len(subscription.DatasourceUids) > 0 && !containsString(subscription.DatasourceUids, event.DatasourceUID) {
		return false
	}
	return labelsMatch(subscription.Labels, event.Labels)
}

func labelsMatch(selector, eventLabels string) bool {
	matched, err := labelsutil.MatchersJSONContained(selector, eventLabels)
	if err != nil {
		return false
	}
	return matched
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

// resolveGroupTargets 把一个通知组展开成实际投递目标：主渠道一次，每个选中的 Webhook 各一次。
func resolveGroupTargets(ctx context.Context, tx *ent.Tx, group *ent.NotificationGroup) ([]notifyTarget, error) {
	channelUIDs, templateUIDs := groupChannelTargets(group)
	targets := make([]notifyTarget, 0, len(channelUIDs))
	seen := map[string]struct{}{}
	for index, channelUID := range channelUIDs {
		if _, ok := seen[channelUID]; ok {
			continue
		}
		if strings.TrimSpace(channelUID) == "" {
			continue
		}
		templateUID := ""
		if index < len(templateUIDs) {
			templateUID = strings.TrimSpace(templateUIDs[index])
		}
		if templateUID == "" {
			continue
		}
		channel, err := tx.NotificationChannel.Query().Where(
			notificationchannel.UIDEQ(channelUID),
			notificationchannel.NamespaceUIDEQ(group.NamespaceUID),
			notificationchannel.DeletedAtIsNil(),
			notificationchannel.StatusEQ(notificationchannel.StatusACTIVE),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		template, err := tx.NotificationTemplate.Query().Where(
			notificationtemplate.UIDEQ(templateUID),
			notificationtemplate.NamespaceUIDEQ(group.NamespaceUID),
			notificationtemplate.DeletedAtIsNil(),
			notificationtemplate.StatusEQ(notificationtemplate.StatusACTIVE),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		if !templateCoversChannel(channel, template) {
			continue
		}
		receivers := ""
		switch channel.Type {
		case notificationchannel.TypeSMTP:
			receivers, err = resolveGroupReceivers(ctx, tx, group, channel.Type)
			if err != nil {
				return nil, err
			}
			if receivers == "" {
				continue
			}
		case notificationchannel.TypeSMS:
			receivers, err = resolveGroupReceivers(ctx, tx, group, channel.Type)
			if err != nil {
				return nil, err
			}
			if receivers == "" {
				continue
			}
		default:
			receivers = notificationReceiver(channel)
		}
		seen[channelUID] = struct{}{}
		targets = append(targets, notifyTarget{
			ChannelUID: channel.UID, TemplateUID: template.UID,
			Sender: notificationSender(channel), Receivers: receivers,
			Subject: template.Subject, Body: template.Body,
		})
	}
	return targets, nil
}

// groupChannelTargets 返回渠道和模板的一一对应列表。历史通知组只有旧版单渠道字段，
// 这里按旧字段展开，升级后无需手工重建通知组。
func groupChannelTargets(group *ent.NotificationGroup) ([]string, []string) {
	if len(group.ChannelUids) > 0 {
		return group.ChannelUids, group.TemplateUids
	}
	if strings.TrimSpace(group.ChannelUID) == "" {
		return nil, nil
	}
	channelUIDs := []string{group.ChannelUID}
	templateUIDs := []string{group.TemplateUID}
	for range group.WebhookChannelUids {
		channelUIDs = append(channelUIDs, "")
		templateUIDs = append(templateUIDs, group.WebhookTemplateUID)
	}
	copy(channelUIDs[1:], group.WebhookChannelUids)
	return channelUIDs, templateUIDs
}

// templateCoversChannel 判断模板是否能覆盖渠道。OTHER Webhook 模板可以覆盖任意应用。
func templateCoversChannel(channel *ent.NotificationChannel, template *ent.NotificationTemplate) bool {
	switch channel.Type {
	case notificationchannel.TypeSMTP:
		return template.Type == notificationtemplate.TypeEMAIL
	case notificationchannel.TypeSMS:
		return template.Type == notificationtemplate.TypeSMS_ALICLOUD
	case notificationchannel.TypeWEBHOOK:
		return template.Type == notificationtemplate.TypeWEBHOOK &&
			(template.App == notificationtemplate.AppOTHER || string(template.App) == string(channel.App))
	default:
		return false
	}
}

// resolveGroupReceivers 按渠道类型合并手填接收方和成员资料。
// 邮件取成员邮箱，短信取成员手机号；同一个人只保留一次。
func resolveGroupReceivers(ctx context.Context, tx *ent.Tx, group *ent.NotificationGroup, kind notificationchannel.Type) (string, error) {
	base := group.EmailReceivers
	memberUIDs := group.EmailMemberUids
	field := "email"
	if kind == notificationchannel.TypeSMS {
		base = group.SmsReceivers
		memberUIDs = group.SmsMemberUids
		field = "phone"
	}
	// 历史数据没有新字段时回退到旧版邮件接收方和成员。
	if kind == notificationchannel.TypeSMTP {
		if base == "" {
			base = group.Receivers
		}
		if len(memberUIDs) == 0 {
			memberUIDs = group.MemberUids
		}
	}
	values := pkgnotify.SplitReceivers(base)
	if len(memberUIDs) > 0 {
		members, err := tx.Member.Query().Where(
			member.UIDIn(memberUIDs...),
			member.NamespaceUIDEQ(group.NamespaceUID),
			member.StatusEQ(member.StatusJOINED),
			member.DeletedAtIsNil(),
		).All(ctx)
		if err != nil {
			return "", err
		}
		userUIDs := make([]string, 0, len(members))
		for _, row := range members {
			userUIDs = append(userUIDs, row.UserUID)
		}
		if len(userUIDs) > 0 {
			users, err := tx.User.Query().Where(
				user.UIDIn(userUIDs...),
				user.StatusEQ(user.StatusACTIVE),
			).All(ctx)
			if err != nil {
				return "", err
			}
			for _, account := range users {
				value := account.Email
				if field == "phone" {
					value = account.Phone
				}
				if strings.TrimSpace(value) != "" {
					values = append(values, value)
				}
			}
		}
	}
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
	return strings.Join(out, ","), nil
}
