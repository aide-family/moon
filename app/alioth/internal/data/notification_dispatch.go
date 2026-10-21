package data

import (
	"context"
	"encoding/json"
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
	selectorMap := map[string]string{}
	if strings.TrimSpace(selector) != "" && selector != "{}" {
		if err := json.Unmarshal([]byte(selector), &selectorMap); err != nil {
			return false
		}
	}
	if len(selectorMap) == 0 {
		return true
	}
	eventMap := map[string]string{}
	if err := json.Unmarshal([]byte(eventLabels), &eventMap); err != nil {
		return false
	}
	for key, value := range selectorMap {
		if eventMap[key] != value {
			return false
		}
	}
	return true
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
	targets := make([]notifyTarget, 0, 2)
	seen := map[string]struct{}{}
	channel, err := tx.NotificationChannel.Query().Where(
		notificationchannel.UIDEQ(group.ChannelUID),
		notificationchannel.NamespaceUIDEQ(group.NamespaceUID),
		notificationchannel.DeletedAtIsNil(),
		notificationchannel.StatusEQ(notificationchannel.StatusACTIVE),
	).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if err == nil {
		template, templateErr := tx.NotificationTemplate.Query().Where(
			notificationtemplate.UIDEQ(group.TemplateUID),
			notificationtemplate.NamespaceUIDEQ(group.NamespaceUID),
			notificationtemplate.DeletedAtIsNil(),
			notificationtemplate.StatusEQ(notificationtemplate.StatusACTIVE),
		).Only(ctx)
		if templateErr != nil && !ent.IsNotFound(templateErr) {
			return nil, templateErr
		}
		if templateErr == nil {
			receivers := ""
			if channel.Type == notificationchannel.TypeSMTP {
				receivers, err = resolveGroupReceivers(ctx, tx, group)
				if err != nil {
					return nil, err
				}
			}
			// SMTP 没有任何有效接收方时只跳过主目标，不影响 Webhook 目标。
			if channel.Type != notificationchannel.TypeSMTP || receivers != "" {
				targets = append(targets, notifyTarget{
					ChannelUID: channel.UID, TemplateUID: template.UID,
					Sender: notificationSender(channel), Receivers: receivers,
					Subject: template.Subject, Body: template.Body,
				})
				seen[channel.UID] = struct{}{}
			}
		}
	}
	if len(group.WebhookChannelUids) == 0 || group.WebhookTemplateUID == "" {
		return targets, nil
	}
	webhookTemplate, err := tx.NotificationTemplate.Query().Where(
		notificationtemplate.UIDEQ(group.WebhookTemplateUID),
		notificationtemplate.NamespaceUIDEQ(group.NamespaceUID),
		notificationtemplate.DeletedAtIsNil(),
		notificationtemplate.StatusEQ(notificationtemplate.StatusACTIVE),
		notificationtemplate.TypeEQ(notificationtemplate.TypeWEBHOOK),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return targets, nil
		}
		return nil, err
	}
	channels, err := tx.NotificationChannel.Query().Where(
		notificationchannel.UIDIn(group.WebhookChannelUids...),
		notificationchannel.NamespaceUIDEQ(group.NamespaceUID),
		notificationchannel.DeletedAtIsNil(),
		notificationchannel.StatusEQ(notificationchannel.StatusACTIVE),
		notificationchannel.TypeEQ(notificationchannel.TypeWEBHOOK),
	).Order(ent.Asc(notificationchannel.FieldUID)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, webhook := range channels {
		if _, ok := seen[webhook.UID]; ok {
			continue
		}
		seen[webhook.UID] = struct{}{}
		targets = append(targets, notifyTarget{
			ChannelUID: webhook.UID, TemplateUID: webhookTemplate.UID,
			Sender: notificationSender(webhook), Receivers: notificationReceiver(webhook),
			Subject: webhookTemplate.Subject, Body: webhookTemplate.Body,
		})
	}
	return targets, nil
}

func resolveGroupReceivers(ctx context.Context, tx *ent.Tx, group *ent.NotificationGroup) (string, error) {
	values := pkgnotify.SplitReceivers(group.Receivers)
	if len(group.MemberUids) > 0 {
		members, err := tx.Member.Query().Where(
			member.UIDIn(group.MemberUids...),
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
				if strings.TrimSpace(account.Email) != "" {
					values = append(values, account.Email)
				}
			}
		}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return strings.Join(out, ","), nil
}
