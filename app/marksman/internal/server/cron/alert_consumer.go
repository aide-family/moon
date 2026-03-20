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

var _ transport.Server = (*ConsumerServer)(nil)

// NewConsumerServer creates a server that consumes alerting jobs from the channel.
func NewConsumerServer(
	alertEventChannelRepo repository.AlertEventChannel,
	alertingEventChannelRepo repository.AlertingEventChannel,
	consumer *biz.AlertEventConsumer,
	helper *klog.Helper,
) *ConsumerServer {
	name := "marksman-cron-alert-consumer"
	opts := []cron.Option{
		cron.WithCronJobChannel(alertingEventChannelRepo.GetJobChannel()),
		cron.WithRemoveJobChannel(alertingEventChannelRepo.GetRemoveJobChannel()),
	}
	return &ConsumerServer{
		alertEventChannelRepo:    alertEventChannelRepo,
		alertingEventChannelRepo: alertingEventChannelRepo,
		consumer:                 consumer,
		helper:                   helper,
		cronServer:               cron.New(name, helper.Logger(), opts...),
	}
}

type ConsumerServer struct {
	alertEventChannelRepo    repository.AlertEventChannel
	alertingEventChannelRepo repository.AlertingEventChannel
	consumer                 *biz.AlertEventConsumer
	helper                   *klog.Helper

	cronServer *cron.Server
}

// Start implements transport.Server.
func (s *ConsumerServer) Start(ctx context.Context) error {
	safety.Go(ctx, "alert-event-consumer", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-s.alertEventChannelRepo.GetChannel():
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
func (s *ConsumerServer) Stop(ctx context.Context) error {
	s.helper.WithContext(ctx).Infow("msg", "alert event consumer stopped")
	return s.cronServer.Stop(ctx)
}
