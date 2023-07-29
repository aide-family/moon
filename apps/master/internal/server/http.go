package server

import (
	"context"
	"github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/swagger-api/openapiv2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	ping "prometheus-manager/api"
	promV1 "prometheus-manager/api/prom/v1"
	crudV1 "prometheus-manager/api/strategy/v1"
	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/apps/master/internal/service"
	promServiceV1 "prometheus-manager/apps/master/internal/service/prom/v1"
	"prometheus-manager/pkg/middler"
	"prometheus-manager/pkg/prom"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	logger log.Logger,
	tp *traceSdk.TracerProvider,
	pingService *service.PingService,
	crudService *service.CrudService,
	dirV1Service *promServiceV1.DirService,
	fileV1Service *promServiceV1.FileService,
	nodeV1Service *promServiceV1.NodeService,
	ruleV1Service *promServiceV1.RuleService,
	groupV1Service *promServiceV1.GroupService,
) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(middler.Cors(), middler.LocalHttpRequestFilter()), // 跨域
		http.Middleware(
			recovery.Recovery(),
			func(handler middleware.Handler) middleware.Handler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					ctx, span := otel.Tracer("http-middleware").Start(ctx, "middleware")
					defer span.End()
					ctx = trace.ContextWithSpanContext(ctx, span.SpanContext())
					return handler(ctx, req)
				}
			},
			logging.Server(logger),
			validate.Validator(),
			tracing.Server(tracing.WithTracerProvider(tp), tracing.WithTracerName("http")),
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
	crudV1.RegisterCrudHTTPServer(srv, crudService)
	promV1.RegisterDirHTTPServer(srv, dirV1Service)
	promV1.RegisterFileHTTPServer(srv, fileV1Service)
	promV1.RegisterNodeHTTPServer(srv, nodeV1Service)
	promV1.RegisterRuleHTTPServer(srv, ruleV1Service)
	promV1.RegisterGroupHTTPServer(srv, groupV1Service)

	log.NewHelper(log.With(logger, "module", "server/http")).Info("http server initialized")

	return srv
}
