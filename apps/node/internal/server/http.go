package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	loadV1 "prometheus-manager/api/strategy/v1/load"
	pullV1 "prometheus-manager/api/strategy/v1/pull"
	pushV1 "prometheus-manager/api/strategy/v1/push"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	pushService *service.PushService,
	loadService *service.LoadService,
	pullService *service.PullService,
	logger log.Logger,
) *http.Server {
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
	pushV1.RegisterPushHTTPServer(srv, pushService)
	pullV1.RegisterPullHTTPServer(srv, pullService)
	loadV1.RegisterLoadHTTPServer(srv, loadService)

	log.NewHelper(log.With(logger, "module", "server/http")).Info("http server initialized")

	return srv
}
