package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/laurel/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/safety"
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
	safety.Go("watchScriptJobEventBus", func() {
		for scriptJob := range c.scriptService.OutScriptJobEventBus() {
			c.AddJobForce(scriptJob)
		}
	}, c.helper.Logger())
	safety.Go("watchRemoveScriptJobEventBus", func() {
		for scriptJob := range c.scriptService.OutRemoveScriptJobEventBus() {
			c.RemoveJob(scriptJob)
		}
	}, c.helper.Logger())
	safety.Go("watchMetricEventBus", func() {
		for metricEvent := range c.scriptService.OutMetricEventBus() {
			c.metricService.PushMetricEvent(context.Background(), metricEvent)
		}
	}, c.helper.Logger())
	for _, job := range c.scriptService.Loads() {
		c.AddJobForce(job)
	}
	return c.CronJobServer.Start(ctx)
}

func (c *CronScriptServer) Stop(ctx context.Context) error {
	return c.CronJobServer.Stop(ctx)
}
