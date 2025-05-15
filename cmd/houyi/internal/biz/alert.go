package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/houyi/internal/conf"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

func NewAlert(
	bc *conf.Bootstrap,
	alertRepo repository.Alert,
	eventBusRepo repository.EventBus,
	logger log.Logger,
) *Alert {
	syncConfig := bc.GetAlert()
	return &Alert{
		helper:       log.NewHelper(log.With(logger, "module", "biz.alert")),
		alertRepo:    alertRepo,
		eventBusRepo: eventBusRepo,
		syncInterval: syncConfig.GetSyncInterval().AsDuration(),
		syncTimeout:  syncConfig.GetSyncTimeout().AsDuration(),
	}
}

type Alert struct {
	helper *log.Helper

	alertRepo    repository.Alert
	eventBusRepo repository.EventBus

	syncInterval time.Duration
	syncTimeout  time.Duration
}

func (a *Alert) Loads() []*server.TickTask {
	return []*server.TickTask{
		{
			Fn:        a.syncAlerts,
			Name:      "syncAlerts",
			Timeout:   a.syncTimeout,
			Interval:  a.syncInterval,
			Immediate: true,
		},
	}
}

func (a *Alert) syncAlerts(ctx context.Context, isStop bool) error {
	if isStop {
		return nil
	}
	alerts, err := a.alertRepo.GetAll(ctx)
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "syncAlerts", "err", err)
		return err
	}
	inAlertEventBus := a.eventBusRepo.InAlertEventBus()
	for _, alert := range alerts {
		inAlertEventBus <- alert
	}
	return nil
}
