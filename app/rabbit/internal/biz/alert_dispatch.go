package biz

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

type alertEmailDispatchPlan struct {
	configUID snowflake.ID
	to        []string
}

func uniqueInt64IDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func recipientGroupEmailRecipients(group *bo.RecipientGroupItemBo) []string {
	if group == nil {
		return nil
	}
	to := make([]string, 0, len(group.Members))
	for _, member := range group.Members {
		if member != nil && member.IsEmail && member.MemberEmail != "" {
			to = appendUnique(to, member.MemberEmail)
		}
	}
	return to
}

func mergeAlertEmailDispatchPlans(plans map[int64]*alertEmailDispatchPlan, configUID snowflake.ID, to []string) {
	if configUID.Int64() == 0 || len(to) == 0 {
		return
	}
	key := configUID.Int64()
	existing, ok := plans[key]
	if !ok {
		plans[key] = &alertEmailDispatchPlan{
			configUID: configUID,
			to:        append([]string(nil), to...),
		}
		return
	}
	for _, addr := range to {
		existing.to = appendUnique(existing.to, addr)
	}
}

func subscriptionEmailRouteKey(subscriptionUID, emailConfigUID snowflake.ID) string {
	return "sub:" + subscriptionUID.String() + ":email:" + emailConfigUID.String()
}

func subscriptionWebhookRouteKey(subscriptionUID, webhookConfigUID snowflake.ID) string {
	return "sub:" + subscriptionUID.String() + ":webhook:" + webhookConfigUID.String()
}

func (b *Alert) dispatchSubscription(ctx context.Context, subscription *bo.AlertSubscriptionItemBo, payload *bo.AlertPayloadBo) error {
	if subscription == nil {
		return nil
	}
	emailPlans := make(map[int64]*alertEmailDispatchPlan)
	webhookConfigUIDs := make(map[int64]struct{})

	for _, recipientGroupUID := range uniqueInt64IDs(subscription.RecipientGroupUIDs) {
		group, err := b.recipientGroupBiz.GetRecipientGroup(ctx, snowflake.ID(recipientGroupUID))
		if err != nil {
			b.helper.WithContext(ctx).Warnw(
				"msg", "dispatch recipient group failed",
				"error", err,
				"recipientGroupUID", recipientGroupUID,
				"subscriptionUID", subscription.UID.Int64(),
			)
			continue
		}
		groupItem := &group.RecipientGroupItemBo
		to := recipientGroupEmailRecipients(groupItem)
		for _, emailConfig := range groupItem.EmailConfigs {
			if emailConfig == nil {
				continue
			}
			mergeAlertEmailDispatchPlans(emailPlans, emailConfig.UID, to)
		}
		for _, webhookConfig := range groupItem.WebhookConfigs {
			if webhookConfig == nil || webhookConfig.UID.Int64() == 0 {
				continue
			}
			webhookConfigUIDs[webhookConfig.UID.Int64()] = struct{}{}
		}
	}

	for _, plan := range emailPlans {
		if plan == nil || len(plan.to) == 0 {
			continue
		}
		routeKey := subscriptionEmailRouteKey(subscription.UID, plan.configUID)
		if !b.shouldDispatchAlertNotification(ctx, payload, routeKey) {
			continue
		}
		if _, err := b.emailBiz.AppendEmailMessage(ctx, &bo.SendEmailBo{
			UID:         plan.configUID,
			Subject:     bo.BuildDefaultAlertSubject(payload),
			Body:        bo.BuildDefaultAlertBody(payload),
			To:          plan.to,
			ContentType: "text/plain",
		}); err != nil {
			return err
		}
		b.markAlertNotificationDispatched(ctx, payload, routeKey)
	}

	for webhookConfigUID := range webhookConfigUIDs {
		routeKey := subscriptionWebhookRouteKey(subscription.UID, snowflake.ID(webhookConfigUID))
		if !b.shouldDispatchAlertNotification(ctx, payload, routeKey) {
			continue
		}
		if _, err := b.webhookBiz.AppendWebhookMessage(ctx, &bo.SendWebhookBo{
			UID:  snowflake.ID(webhookConfigUID),
			Data: bo.BuildAlertTemplateData(payload),
		}); err != nil {
			return err
		}
		b.markAlertNotificationDispatched(ctx, payload, routeKey)
	}

	return b.dispatchSubscriptionMembers(ctx, subscription, payload)
}
