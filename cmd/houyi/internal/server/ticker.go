package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/conf"
	"github.com/aide-family/moon/cmd/houyi/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server"
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
