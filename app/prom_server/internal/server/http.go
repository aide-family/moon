package server

import (
	"prometheus-manager/api/dict"
	"prometheus-manager/api/ping"
	"prometheus-manager/api/prom/strategy"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/service"
	"prometheus-manager/app/prom_server/internal/service/dictservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	pingService *service.PingService,
	dictService *dictservice.Service,
	strategyService *promservice.StrategyService,
	logger log.Logger,
) *http.Server {
	logHelper := log.NewHelper(log.With(logger, "module", "http"))
	defer logHelper.Info("NewHTTPServer done")
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	ping.RegisterPingHTTPServer(srv, pingService)
	dict.RegisterDictHTTPServer(srv, dictService)
	strategy.RegisterStrategyHTTPServer(srv, strategyService)

	return srv
}
