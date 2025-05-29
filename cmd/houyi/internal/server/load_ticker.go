package server

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/service"
	"github.com/aide-family/moon/pkg/plugin/server/ticker_server"
)

type LoadTickerServer struct {
	*ticker_server.Tickers
}

func NewLoadTickerServer(loadService *service.LoadService, logger log.Logger) *LoadTickerServer {
	return &LoadTickerServer{
		Tickers: ticker_server.NewTickers(
			ticker_server.WithTickersTasks(loadService.Loads()...),
			ticker_server.WithTickersLogger(logger),
		),
	}
}
