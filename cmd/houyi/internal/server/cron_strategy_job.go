package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/moon/cmd/houyi/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/safety"
)

var _ transport.Server = (*CronStrategyJobServer)(nil)

func NewCronStrategyJobServer(evaluateService *service.EventBusService, logger log.Logger) *CronStrategyJobServer {
	return &CronStrategyJobServer{
		evaluateService: evaluateService,
		helper:          log.NewHelper(log.With(logger, "module", "server.cron.strategy.job")),
		CronJobServer:   cron_server.NewCronJobServer("Strategy", logger),
	}
}

type CronStrategyJobServer struct {
	evaluateService *service.EventBusService
	helper          *log.Helper
	*cron_server.CronJobServer
}

func (c *CronStrategyJobServer) Start(ctx context.Context) error {
	safety.Go("watchStrategyJobEventBus", func() {
		for strategyJob := range c.evaluateService.OutStrategyJobEventBus() {
			if strategyJob.GetEnable() {
				c.AddJobForce(strategyJob)
			} else {
				c.RemoveJob(strategyJob)
			}
		}
	}, c.helper.Logger())
	return c.CronJobServer.Start(ctx)
}

func (c *CronStrategyJobServer) Stop(ctx context.Context) error {
	return c.CronJobServer.Stop(ctx)
}
