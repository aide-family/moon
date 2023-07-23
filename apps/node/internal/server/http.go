package server

import (
	"github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	ping "prometheus-manager/api"
	loadV1 "prometheus-manager/api/strategy/v1/load"
	pullV1 "prometheus-manager/api/strategy/v1/pull"
	pushV1 "prometheus-manager/api/strategy/v1/push"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
	"prometheus-manager/pkg/middler"
	"prometheus-manager/pkg/prom"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	logger log.Logger,
	tp *traceSdk.TracerProvider,
	pingService *service.PingService,
	pushService *service.PushService,
	loadService *service.LoadService,
	pullService *service.PullService,
) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(middler.Cors(), middler.LocalHttpRequestFilter()), // 跨域
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			validate.Validator(),
			tracing.Server(tracing.WithTracerProvider(tp)),
			ratelimit.Server(),
			metrics.Server(
				metrics.WithSeconds(prometheus.NewHistogram(prom.MetricSeconds)),
				metrics.WithRequests(prometheus.NewCounter(prom.MetricRequests)),
			),
			middler.IpMetric(prom.IpMetricCounter),
			validate.Validator(),
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

	// swagger api
	srv.HandlePrefix("/q/", openapiv2.NewHandler())
	// prometheus metrics
	srv.HandlePrefix("/metrics", promhttp.Handler())

	ping.RegisterPingHTTPServer(srv, pingService)
	pushV1.RegisterPushHTTPServer(srv, pushService)
	pullV1.RegisterPullHTTPServer(srv, pullService)
	loadV1.RegisterLoadHTTPServer(srv, loadService)

	log.NewHelper(log.With(logger, "module", "server/http")).Info("http server initialized")

	return srv
}
