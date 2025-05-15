package server

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/houyi/internal/service"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

type LoadTickerServer struct {
	*server.Tickers
}

func NewLoadTickerServer(loadService *service.LoadService, logger log.Logger) *LoadTickerServer {
	return &LoadTickerServer{
		Tickers: server.NewTickers(
			server.WithTickersTasks(loadService.Loads()...),
			server.WithTickersLogger(logger),
		),
	}
}
