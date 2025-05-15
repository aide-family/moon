package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/cmd/houyi/internal/service"
)

func NewEventBusServer(
	eventBusService *service.EventBusService,
	alertService *service.AlertService,
	logger log.Logger,
) *EventBusServer {
	return &EventBusServer{
		helper:          log.NewHelper(logger),
		stop:            make(chan struct{}),
		eventBusService: eventBusService,
		alertService:    alertService,
	}
}

type EventBusServer struct {
	helper *log.Helper
	stop   chan struct{}

	eventBusService *service.EventBusService
	alertService    *service.AlertService
}

func (e *EventBusServer) Start(ctx context.Context) error {
	defer e.helper.WithContext(ctx).Info("[EventBus] server is started")
	go func() {
		defer func() {
			if err := recover(); err != nil {
				e.helper.Errorw("method", "watchEventBus", "panic", err)
			}
		}()
		for {
			select {
			case <-e.stop:
				return
			case <-ctx.Done():
				return
			case alert := <-e.eventBusService.OutAlertEventBus():
				e.helper.Debugw("msg", "[EventBus] receive alert event", "alert", alert.GetFingerprint())
				e.alertService.InnerPush(ctx, alert)
			}
		}
	}()

	return nil
}

func (e *EventBusServer) Stop(ctx context.Context) error {
	defer e.helper.WithContext(ctx).Info("[EventBus] server is stopped")
	close(e.stop)
	return nil
}
