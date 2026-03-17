// Package cron is the cron server for the marksman.
package cron

import (
	"context"

	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/server/cron"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/repository"
)

var _ transport.Server = (*AlertEventConsumerServer)(nil)

// NewAlertEventConsumerServer creates a server that consumes alert events from the channel.
func NewAlertEventConsumerServer(
	alertEventChannel repository.AlertEventChannel,
	alertingRepo repository.Alerting,
	consumer *biz.AlertEventConsumer,
	helper *klog.Helper,
) *AlertEventConsumerServer {
	name := "marksman-cron-alerting-consumer"
	opts := []cron.Option{
		cron.WithCronJobChannel(alertingRepo.GetJobChannel()),
		cron.WithRemoveJobChannel(alertingRepo.GetRemoveJobChannel()),
	}
	return &AlertEventConsumerServer{
		alertCh:      alertEventChannel,
		alertingRepo: alertingRepo,
		consumer:     consumer,
		helper:       helper,
		cronServer:   cron.New(name, helper.Logger(), opts...),
	}
}

type AlertEventConsumerServer struct {
	alertCh      repository.AlertEventChannel
	alertingRepo repository.Alerting
	consumer     *biz.AlertEventConsumer
	helper       *klog.Helper

	cronServer *cron.Server
}

// Start implements transport.Server.
func (s *AlertEventConsumerServer) Start(ctx context.Context) error {
	ch := s.alertCh.GetChannel()
	safety.Go(ctx, "alert-event-consumer", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-ch:
				if !ok {
					return nil
				}
				s.consumer.Handle(ctx, event)
			}
		}
	})
	safety.Go(ctx, "alerting-consumer", func(ctx context.Context) error {
		return s.cronServer.Start(ctx)
	})
	s.helper.WithContext(ctx).Infow("msg", "alert event consumer started")
	return nil
}

// Stop implements transport.Server.
func (s *AlertEventConsumerServer) Stop(ctx context.Context) error {
	s.helper.WithContext(ctx).Infow("msg", "alert event consumer stopped")
	return s.cronServer.Stop(ctx)
}
