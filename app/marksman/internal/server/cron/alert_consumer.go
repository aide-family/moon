// Package cron is the cron server for the marksman.
package cron

import (
	"context"
	"sync"

	"github.com/aide-family/magicbox/safety"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/repository"
)

var _ transport.Server = (*AlertEventConsumerServer)(nil)

// NewAlertEventConsumerServer creates a server that consumes alert events from the channel.
func NewAlertEventConsumerServer(
	alertEventChannel repository.AlertEventChannel,
	consumer *biz.AlertEventConsumer,
	helper *klog.Helper,
) *AlertEventConsumerServer {
	return &AlertEventConsumerServer{
		alertCh: alertEventChannel,
		consumer: consumer,
		helper:   helper,
	}
}

type AlertEventConsumerServer struct {
	alertCh  repository.AlertEventChannel
	consumer *biz.AlertEventConsumer
	helper   *klog.Helper
	stop     func()
	stopped  sync.WaitGroup
}

// Start implements transport.Server.
func (s *AlertEventConsumerServer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.stop = cancel
	s.stopped.Add(1)
	ch := s.alertCh.GetChannel()
	safety.Go(ctx, "alert-event-consumer", func(ctx context.Context) error {
		defer s.stopped.Done()
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
	s.helper.WithContext(ctx).Infow("msg", "alert event consumer started")
	return nil
}

// Stop implements transport.Server.
func (s *AlertEventConsumerServer) Stop(ctx context.Context) error {
	if s.stop != nil {
		s.stop()
	}
	s.stopped.Wait()
	s.helper.WithContext(ctx).Infow("msg", "alert event consumer stopped")
	return nil
}
