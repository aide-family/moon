package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/laurel/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

type CronScriptServer struct {
	helper        *log.Helper
	scriptService *service.ScriptService
	metricService *service.MetricService
	*cron_server.CronJobServer
}

func NewCronScriptServer(
	scriptService *service.ScriptService,
	metricService *service.MetricService,
	logger log.Logger,
) *CronScriptServer {
	return &CronScriptServer{
		helper:        log.NewHelper(log.With(logger, "module", "server.cron.script")),
		scriptService: scriptService,
		metricService: metricService,
		CronJobServer: cron_server.NewCronJobServer("Script.Job", logger),
	}
}

func (c *CronScriptServer) Start(ctx context.Context) error {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.helper.Errorw("method", "watchEventBus", "panic", err)
			}
		}()
		for scriptJob := range c.scriptService.OutScriptJobEventBus() {
			c.AddJob(scriptJob)
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.helper.Errorw("method", "watchEventBus", "panic", err)
			}
		}()
		for scriptJob := range c.scriptService.OutRemoveScriptJobEventBus() {
			c.RemoveJob(scriptJob)
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.helper.Errorw("method", "watchEventBus", "panic", err)
			}
		}()
		for metricEvent := range c.scriptService.OutMetricEventBus() {
			c.metricService.PushMetricEvent(context.Background(), metricEvent)
		}
	}()
	for _, job := range c.scriptService.Loads() {
		c.AddJobForce(job)
	}
	return c.CronJobServer.Start(ctx)
}

func (c *CronScriptServer) Stop(ctx context.Context) error {
	return c.CronJobServer.Stop(ctx)
}
