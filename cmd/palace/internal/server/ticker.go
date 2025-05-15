package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/moon-monitor/moon/cmd/palace/internal/service"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

var _ transport.Server = (*TickerServer)(nil)

func NewTickerServer(loadService *service.LoadService, logger log.Logger) *TickerServer {
	cronServer := server.NewCronJobServer("palace.Ticker", logger, loadService.LoadJobs()...)
	return &TickerServer{cronServer}
}

type TickerServer struct {
	*server.CronJobServer
}
