package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/moon/cmd/houyi/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/safety"
)

var _ transport.Server = (*CronAlertJobServer)(nil)

func NewCronAlertJobServer(
	evaluateService *service.EventBusService,
	alertService *service.AlertService,
	logger log.Logger,
) *CronAlertJobServer {
	return &CronAlertJobServer{
		evaluateService: evaluateService,
		alertService:    alertService,
		helper:          log.NewHelper(log.With(logger, "module", "server.cron.alert.job")),
		CronJobServer:   cron_server.NewCronJobServer("Alert", logger),
	}
}

type CronAlertJobServer struct {
	evaluateService *service.EventBusService
	alertService    *service.AlertService

	helper *log.Helper
	*cron_server.CronJobServer
}

func (c *CronAlertJobServer) Start(ctx context.Context) error {
	safety.Go("watchAlertJobEventBus", func() {
		for alertJob := range c.evaluateService.OutAlertJobEventBus() {
			if alertJob.GetAlert().IsResolved() {
				c.helper.Debugw("method", "watchEventBus", "alertJobResolved", alertJob.GetAlert().GetFingerprint())
				c.RemoveJob(alertJob)
				continue
			}
			c.AddJob(alertJob)
		}
	}, c.helper.Logger())
	return c.CronJobServer.Start(ctx)
}

func (c *CronAlertJobServer) Stop(ctx context.Context) error {
	return c.CronJobServer.Stop(ctx)
}
