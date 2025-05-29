package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/conf"
	"github.com/aide-family/moon/cmd/houyi/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/ticker_server"
)

func NewTicker(bc *conf.Bootstrap, healthService *service.HealthService, logger log.Logger) *ticker_server.Ticker {
	serverConfig := bc.GetServer()
	microConfig := bc.GetPalace()
	return ticker_server.NewTicker(serverConfig.GetOnlineInterval().AsDuration(), &ticker_server.TickTask{
		Name:    "health.register.houyi",
		Timeout: microConfig.GetTimeout().AsDuration(),
		Fn: func(ctx context.Context, isStop bool) error {
			return healthService.Register(ctx, !isStop)
		},
	}, ticker_server.WithTickerLogger(logger))
}
