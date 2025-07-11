package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/moon/cmd/palace/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

var _ transport.Server = (*TickerServer)(nil)

func NewTickerServer(loadService *service.LoadService, logger log.Logger) *TickerServer {
	cronServer := cron_server.NewCronJobServer("palace.CronJob", logger, loadService.LoadJobs()...)
	return &TickerServer{cronServer}
}

type TickerServer struct {
	*cron_server.CronJobServer
}
