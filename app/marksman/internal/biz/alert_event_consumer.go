package biz

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

// AlertEventConsumer consumes alert events from the channel (e.g. log, persist, or send to rabbit).
type AlertEventConsumer struct {
	helper         *klog.Helper
	alertEventRepo repository.AlertEvent
	strategyRepo   repository.Strategy
}

// NewAlertEventConsumer creates an AlertEventConsumer.
func NewAlertEventConsumer(
	helper *klog.Helper,
	alertEventRepo repository.AlertEvent,
	strategyRepo repository.Strategy,
) *AlertEventConsumer {
	return &AlertEventConsumer{
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "alert_event_consumer")),
		alertEventRepo: alertEventRepo,
		strategyRepo:   strategyRepo,
	}
}

// Handle processes one alert event: persists to DB and logs.
func (c *AlertEventConsumer) Handle(ctx context.Context, event *bo.AlertEventBo) {
	if event == nil {
		return
	}
	ctx = contextx.WithNamespace(ctx, event.NamespaceUID)
	strategy, err := c.strategyRepo.GetStrategy(ctx, event.StrategyUID)
	if err != nil {
		c.helper.WithContext(ctx).Errorw("msg", "get strategy for alert event failed", "error", err, "strategyUID", event.StrategyUID.Int64())
		return
	}
	if err := c.alertEventRepo.CreateAlertEvent(ctx, event, strategy.StrategyGroupUID); err != nil {
		c.helper.WithContext(ctx).Errorw("msg", "create alert event failed", "error", err, "strategyUID", event.StrategyUID.Int64())
		return
	}
	levelName := ""
	if event.Level != nil {
		levelName = event.Level.Name
	}
	c.helper.WithContext(ctx).Infow(
		"msg", "alert event persisted",
		"strategyUID", event.StrategyUID.Int64(),
		"namespaceUID", event.NamespaceUID.Int64(),
		"level", levelName,
		"summary", event.Summary,
		"value", event.Value,
		"firedAt", event.FiredAt,
		"expr", event.Expr,
	)
}
