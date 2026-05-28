package biz

import (
	"context"

	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

type RabbitAlertPusher struct {
	rabbitAlertRepo repository.RabbitAlert
	dedup           repository.AlertPushDedup
	helper          *klog.Helper
}

func NewRabbitAlertPusher(
	rabbitAlertRepo repository.RabbitAlert,
	dedup repository.AlertPushDedup,
	helper *klog.Helper,
) *RabbitAlertPusher {
	return &RabbitAlertPusher{
		rabbitAlertRepo: rabbitAlertRepo,
		dedup:           dedup,
		helper:          klog.NewHelper(klog.With(helper.Logger(), "biz", "rabbit_alert_pusher")),
	}
}

func (p *RabbitAlertPusher) PushFiringAlert(ctx context.Context, alertEventUID snowflake.ID, event *bo.AlertEventBo) {
	if p == nil || event == nil {
		return
	}
	if event.Fingerprint == "" {
		p.pushFiring(ctx, alertEventUID, event)
		return
	}
	exists, err := p.dedup.ExistsFiringPush(ctx, event.NamespaceUID, event.Fingerprint)
	if err != nil {
		p.helper.WithContext(ctx).Warnw(
			"msg", "check firing alert push dedup failed",
			"error", err,
			"fingerprint", event.Fingerprint,
			"alertEventUID", alertEventUID.Int64(),
		)
	} else if exists {
		p.helper.WithContext(ctx).Debugw(
			"msg", "skip duplicate firing alert push within dedup window",
			"fingerprint", event.Fingerprint,
			"alertEventUID", alertEventUID.Int64(),
		)
		return
	}
	if !p.pushFiring(ctx, alertEventUID, event) {
		return
	}
	if err := p.dedup.MarkFiringPush(ctx, event.NamespaceUID, event.Fingerprint); err != nil {
		p.helper.WithContext(ctx).Warnw(
			"msg", "mark firing alert push dedup failed",
			"error", err,
			"fingerprint", event.Fingerprint,
			"alertEventUID", alertEventUID.Int64(),
		)
	}
}

func (p *RabbitAlertPusher) PushRecoveredAlert(ctx context.Context, item *bo.AlertEventItemBo) error {
	if p == nil || item == nil {
		return nil
	}
	fingerprint := item.Fingerprint
	if fingerprint == "" {
		fingerprint = bo.BuildRecoveredFingerprint(item)
	}
	if fingerprint != "" {
		if err := p.dedup.ClearFiringPush(ctx, item.NamespaceUID, fingerprint); err != nil {
			p.helper.WithContext(ctx).Warnw(
				"msg", "clear firing alert push dedup failed",
				"error", err,
				"fingerprint", fingerprint,
				"alertEventUID", item.UID.Int64(),
			)
		}
	}
	req := bo.NewRabbitReceivePrometheusWebhookRequestFromAlertEventItem(item)
	if req == nil {
		return nil
	}
	_, err := p.rabbitAlertRepo.ReceivePrometheusWebhook(ctx, req)
	return err
}

func (p *RabbitAlertPusher) pushFiring(ctx context.Context, alertEventUID snowflake.ID, event *bo.AlertEventBo) bool {
	req := bo.NewRabbitReceivePrometheusWebhookRequestFromAlertEvent(alertEventUID, event)
	if req == nil {
		return false
	}
	if _, err := p.rabbitAlertRepo.ReceivePrometheusWebhook(ctx, req); err != nil {
		p.helper.WithContext(ctx).Warnw(
			"msg", "push firing alert to rabbit failed",
			"error", err,
			"alertEventUID", alertEventUID.Int64(),
		)
		return false
	}
	return true
}
