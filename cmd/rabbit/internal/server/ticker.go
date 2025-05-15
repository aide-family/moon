package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/rabbit/internal/conf"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/service"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

func NewTicker(bc *conf.Bootstrap, healthService *service.HealthService, logger log.Logger) *server.Ticker {
	serverConfig := bc.GetServer()
	microConfig := bc.GetPalace()
	return server.NewTicker(serverConfig.GetOnlineInterval().AsDuration(), &server.TickTask{
		Name:    "health.Online",
		Timeout: microConfig.GetTimeout().AsDuration(),
		Fn: func(ctx context.Context, isStop bool) error {
			if isStop {
				return healthService.Offline(ctx)
			}
			return healthService.Online(ctx)
		},
	}, server.WithTickerLogger(logger))
}
