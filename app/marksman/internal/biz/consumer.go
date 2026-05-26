package biz

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/evaluator"
	"github.com/aide-family/marksman/internal/biz/repository"
)

// AlertEventConsumer consumes alert events from the channel (e.g. log, persist, or send to rabbit).
type AlertEventConsumer struct {
	helper                   *klog.Helper
	alertEventRepo           repository.AlertEvent
	rabbitAlertRepo          repository.RabbitAlert
	strategyRepo             repository.Strategy
	alertingEventChannelRepo repository.AlertingEventChannel
}

// NewAlertEventConsumer creates an AlertEventConsumer.
func NewAlertEventConsumer(
	helper *klog.Helper,
	alertEventRepo repository.AlertEvent,
	rabbitAlertRepo repository.RabbitAlert,
	strategyRepo repository.Strategy,
	alertingEventChannelRepo repository.AlertingEventChannel,
) *AlertEventConsumer {
	return &AlertEventConsumer{
		helper:                   klog.NewHelper(klog.With(helper.Logger(), "biz", "alert_event_consumer")),
		alertEventRepo:           alertEventRepo,
		rabbitAlertRepo:          rabbitAlertRepo,
		strategyRepo:             strategyRepo,
		alertingEventChannelRepo: alertingEventChannelRepo,
	}
}

// Handle processes one alert event: persists to DB and logs.
func (c *AlertEventConsumer) Handle(ctx context.Context, event *bo.AlertEventBo) {
	if event == nil {
		return
	}
	ctx = contextx.WithNamespace(ctx, event.NamespaceUID)
	alertEventUID, err := c.alertEventRepo.SaveAlertEvent(ctx, event)
	if err != nil {
		c.helper.WithContext(ctx).Errorw("msg", "create alert event failed", "error", err, "strategyUID", event.StrategyUID.Int64())
		return
	}
	if req := bo.NewRabbitReceivePrometheusWebhookRequestFromAlertEvent(alertEventUID, event); req != nil {
		if _, err := c.rabbitAlertRepo.ReceivePrometheusWebhook(ctx, req); err != nil {
			c.helper.WithContext(ctx).Warnw("msg", "push firing alert to rabbit failed", "error", err, "alertEventUID", alertEventUID.Int64())
		}
	}

	c.helper.WithContext(ctx).Debugw("msg", "alert event persisted", "event", event)
	alerting := evaluator.NewAlerting(alertEventUID, event, c.alertEventRepo, c.alertingEventChannelRepo, func(cbCtx context.Context, uid snowflake.ID) error {
		alertEvent, err := c.alertEventRepo.GetAlertEvent(cbCtx, uid)
		if err != nil {
			return err
		}
		req := bo.NewRabbitReceivePrometheusWebhookRequestFromAlertEventItem(alertEvent)
		if req == nil {
			return nil
		}
		_, err = c.rabbitAlertRepo.ReceivePrometheusWebhook(cbCtx, req)
		return err
	})
	c.alertingEventChannelRepo.Append(alerting)
}
