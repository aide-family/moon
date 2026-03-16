package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
)

// AlertEventConsumer consumes alert events from the channel (e.g. log, persist, or send to rabbit).
type AlertEventConsumer struct {
	helper *klog.Helper
}

// NewAlertEventConsumer creates an AlertEventConsumer.
func NewAlertEventConsumer(helper *klog.Helper) *AlertEventConsumer {
	return &AlertEventConsumer{
		helper: klog.NewHelper(klog.With(helper.Logger(), "biz", "alert_event_consumer")),
	}
}

// Handle processes one alert event. Currently logs; can be extended to persist or notify.
func (c *AlertEventConsumer) Handle(ctx context.Context, event *bo.AlertEventBo) {
	if event == nil {
		return
	}
	levelName := ""
	if event.Level != nil {
		levelName = event.Level.Name
	}
	c.helper.WithContext(ctx).Infow(
		"msg", "alert event",
		"strategyUID", event.StrategyUID.Int64(),
		"namespaceUID", event.NamespaceUID.Int64(),
		"level", levelName,
		"summary", event.Summary,
		"value", event.Value,
		"firedAt", event.FiredAt,
		"expr", event.Expr,
	)
}
