package biz

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/evaluator"
	"github.com/aide-family/marksman/internal/biz/repository"
)

// AlertEventConsumer consumes alert events from the channel (e.g. log, persist, or send to rabbit).
type AlertEventConsumer struct {
	helper                   *klog.Helper
	alertEventRepo           repository.AlertEvent
	strategyRepo             repository.Strategy
	alertingEventChannelRepo repository.AlertingEventChannel
}

// NewAlertEventConsumer creates an AlertEventConsumer.
func NewAlertEventConsumer(
	helper *klog.Helper,
	alertEventRepo repository.AlertEvent,
	strategyRepo repository.Strategy,
	alertingEventChannelRepo repository.AlertingEventChannel,
) *AlertEventConsumer {
	return &AlertEventConsumer{
		helper:                   klog.NewHelper(klog.With(helper.Logger(), "biz", "alert_event_consumer")),
		alertEventRepo:           alertEventRepo,
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

	c.helper.WithContext(ctx).Debugw("msg", "alert event persisted", "event", event)
	alerting := evaluator.NewAlerting(alertEventUID, event, c.alertEventRepo, c.alertingEventChannelRepo)
	c.alertingEventChannelRepo.Append(alerting)
}
