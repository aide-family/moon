package biz

import (
	"context"
	"slices"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/plugin/cache"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewAlert(
	alertSubscriptionRepo repository.AlertSubscription,
	memberRepo repository.Member,
	recipientGroupBiz *RecipientGroup,
	emailBiz *Email,
	webhookBiz *Webhook,
	cache cache.Interface,
	helper *klog.Helper,
) *Alert {
	return &Alert{
		alertSubscriptionRepo: alertSubscriptionRepo,
		memberRepo:            memberRepo,
		recipientGroupBiz:     recipientGroupBiz,
		emailBiz:              emailBiz,
		webhookBiz:            webhookBiz,
		cache:                 cache,
		helper:                klog.NewHelper(klog.With(helper.Logger(), "biz", "alert")),
	}
}

type Alert struct {
	alertSubscriptionRepo repository.AlertSubscription
	memberRepo            repository.Member
	recipientGroupBiz     *RecipientGroup
	emailBiz              *Email
	webhookBiz            *Webhook
	cache                 cache.Interface
	helper                *klog.Helper
}

func (b *Alert) ReceivePrometheusWebhook(ctx context.Context, req *bo.ReceivePrometheusWebhookBo) (int64, error) {
	namespaceUID, err := req.NamespaceUID()
	if err != nil {
		return 0, err
	}
	ctx = contextx.WithNamespace(ctx, namespaceUID)

	var total int64
	for _, alert := range req.Alerts {
		if alert == nil {
			continue
		}
		payload := bo.NewAlertPayloadBo(req, alert)
		if payload == nil {
			continue
		}
		if err := b.dispatchAlert(ctx, payload); err != nil {
			b.helper.WithContext(ctx).Errorw("msg", "dispatch alert failed", "error", err, "fingerprint", alert.Fingerprint)
			return total, err
		}
		total++
	}
	return total, nil
}

func (b *Alert) CreateAlertSubscription(ctx context.Context, req *bo.CreateAlertSubscriptionBo) (snowflake.ID, error) {
	if len(req.Labels) == 0 {
		return 0, merr.ErrorParams("labels are required")
	}
	if subscription, err := b.alertSubscriptionRepo.GetAlertSubscriptionByName(ctx, req.Name); err == nil {
		return 0, merr.ErrorParams("alert subscription %s already exists, uid: %d", req.Name, subscription.UID.Int64())
	} else if !merr.IsNotFound(err) {
		return 0, err
	}
	return b.alertSubscriptionRepo.CreateAlertSubscription(ctx, req)
}

func (b *Alert) UpdateAlertSubscription(ctx context.Context, req *bo.UpdateAlertSubscriptionBo) error {
	if len(req.Labels) == 0 {
		return merr.ErrorParams("labels are required")
	}
	return b.alertSubscriptionRepo.UpdateAlertSubscription(ctx, req)
}

func (b *Alert) DeleteAlertSubscription(ctx context.Context, uid snowflake.ID) error {
	return b.alertSubscriptionRepo.DeleteAlertSubscription(ctx, uid)
}

func (b *Alert) GetAlertSubscription(ctx context.Context, uid snowflake.ID) (*bo.AlertSubscriptionDetailBo, error) {
	subscription, err := b.alertSubscriptionRepo.GetAlertSubscription(ctx, uid)
	if err != nil {
		return nil, err
	}
	if err := b.fillSubscriptionDetailMembers(ctx, subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (b *Alert) ListAlertSubscription(ctx context.Context, req *bo.ListAlertSubscriptionBo) (*bo.PageResponseBo[*bo.AlertSubscriptionItemBo], error) {
	page, err := b.alertSubscriptionRepo.ListAlertSubscription(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := b.fillSubscriptionMembers(ctx, page.GetItems()); err != nil {
		return nil, err
	}
	return page, nil
}

func (b *Alert) UpdateAlertSubscriptionStatus(ctx context.Context, req *bo.UpdateAlertSubscriptionStatusBo) error {
	return b.alertSubscriptionRepo.UpdateAlertSubscriptionStatus(ctx, req)
}

func (b *Alert) dispatchAlert(ctx context.Context, payload *bo.AlertPayloadBo) error {
	subscriptions, err := b.alertSubscriptionRepo.ListEnabledAlertSubscriptions(ctx)
	if err != nil {
		return err
	}
	if err := b.fillSubscriptionMembers(ctx, subscriptions); err != nil {
		return err
	}
	for _, subscription := range subscriptions {
		if subscription == nil || !subscription.MatchesLabels(payload.Labels) {
			continue
		}
		if err := b.dispatchSubscription(ctx, subscription, payload); err != nil {
			b.helper.WithContext(ctx).Warnw("msg", "dispatch alert subscription failed", "error", err, "subscriptionUID", subscription.UID.Int64())
		}
	}
	return nil
}

func (b *Alert) dispatchSubscriptionMembers(ctx context.Context, subscription *bo.AlertSubscriptionItemBo, payload *bo.AlertPayloadBo) error {
	if subscription == nil || !subscription.DirectEmailEnabled() {
		return nil
	}
	to := make([]string, 0, len(subscription.Members))
	for _, member := range subscription.Members {
		if member != nil && member.IsEmail && member.MemberEmail != "" {
			to = appendUnique(to, member.MemberEmail)
		}
	}
	if len(to) == 0 {
		b.helper.WithContext(ctx).Warnw(
			"msg", "skip direct member email: no recipient addresses",
			"subscriptionUID", subscription.UID.Int64(),
			"subscriptionName", subscription.Name,
		)
		return nil
	}
	routeKey := subscriptionEmailRouteKey(subscription.UID, subscription.DirectMemberEmailConfigUID)
	if !b.shouldDispatchAlertNotification(ctx, payload, routeKey) {
		return nil
	}
	if subscription.DirectMemberTemplateUID > 0 {
		_, err := b.emailBiz.AppendEmailMessageWithTemplate(ctx, &bo.SendEmailWithTemplateBo{
			UID:         subscription.DirectMemberEmailConfigUID,
			TemplateUID: subscription.DirectMemberTemplateUID,
			JSONData:    []byte(bo.BuildAlertTemplateData(payload)),
			To:          to,
		})
		if err != nil {
			return err
		}
		b.markAlertNotificationDispatched(ctx, payload, routeKey)
		return nil
	}
	_, err := b.emailBiz.AppendEmailMessage(ctx, &bo.SendEmailBo{
		UID:         subscription.DirectMemberEmailConfigUID,
		Subject:     bo.BuildDefaultAlertSubject(payload),
		Body:        bo.BuildDefaultAlertBody(payload),
		To:          to,
		ContentType: "text/plain",
	})
	if err != nil {
		return err
	}
	b.markAlertNotificationDispatched(ctx, payload, routeKey)
	return nil
}

func (b *Alert) fillSubscriptionMembers(ctx context.Context, subscriptions []*bo.AlertSubscriptionItemBo) error {
	for _, subscription := range subscriptions {
		if subscription == nil {
			continue
		}
		if err := fillNotificationMemberDetails(ctx, b.memberRepo, b.helper, subscription.Members); err != nil {
			return err
		}
	}
	return nil
}

func (b *Alert) fillSubscriptionDetailMembers(ctx context.Context, detail *bo.AlertSubscriptionDetailBo) error {
	if detail == nil {
		return nil
	}
	if err := fillNotificationMemberDetails(ctx, b.memberRepo, b.helper, detail.Members); err != nil {
		return err
	}
	for _, group := range detail.RecipientGroups {
		if group == nil {
			continue
		}
		if err := fillNotificationMemberDetails(ctx, b.memberRepo, b.helper, group.Members); err != nil {
			return err
		}
	}
	return nil
}

func appendUnique(items []string, value string) []string {
	if value == "" {
		return items
	}
	if slices.Contains(items, value) {
		return items
	}
	return append(items, value)
}
