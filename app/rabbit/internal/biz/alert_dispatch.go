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

type alertWebhookDispatchPlan struct {
	configUID   snowflake.ID
	templateUID snowflake.ID
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

func mergeAlertWebhookDispatchPlans(plans map[int64]*alertWebhookDispatchPlan, configUID, templateUID snowflake.ID) {
	if configUID.Int64() == 0 {
		return
	}
	key := configUID.Int64()
	existing, ok := plans[key]
	if !ok {
		plans[key] = &alertWebhookDispatchPlan{
			configUID:   configUID,
			templateUID: templateUID,
		}
		return
	}
	if existing.templateUID.Int64() == 0 && templateUID.Int64() > 0 {
		existing.templateUID = templateUID
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
	webhookPlans := make(map[int64]*alertWebhookDispatchPlan)

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
			templateUID := bo.MatchWebhookTemplate(groupItem.Templates, webhookConfig.App)
			mergeAlertWebhookDispatchPlans(webhookPlans, webhookConfig.UID, templateUID)
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

	templateJSONData := []byte(bo.BuildAlertTemplateData(payload))
	for _, plan := range webhookPlans {
		if plan == nil || plan.configUID.Int64() == 0 {
			continue
		}
		routeKey := subscriptionWebhookRouteKey(subscription.UID, plan.configUID)
		if !b.shouldDispatchAlertNotification(ctx, payload, routeKey) {
			continue
		}
		var err error
		if plan.templateUID.Int64() > 0 {
			_, err = b.webhookBiz.AppendWebhookMessageWithTemplate(ctx, &bo.SendWebhookWithTemplateBo{
				UID:         plan.configUID,
				TemplateUID: plan.templateUID,
				JSONData:    templateJSONData,
			})
		} else {
			_, err = b.webhookBiz.AppendWebhookMessage(ctx, &bo.SendWebhookBo{
				UID:  plan.configUID,
				Data: string(templateJSONData),
			})
		}
		if err != nil {
			return err
		}
		b.markAlertNotificationDispatched(ctx, payload, routeKey)
	}

	return b.dispatchSubscriptionMembers(ctx, subscription, payload)
}
