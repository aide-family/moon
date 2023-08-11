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
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	ping "prometheus-manager/api"
	promV1 "prometheus-manager/api/prom/v1"
	"prometheus-manager/pkg/middler"
	"prometheus-manager/pkg/prom"

	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/apps/master/internal/service"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server,
	logger log.Logger,
	tp *traceSdk.TracerProvider,
	pingService *service.PingService,
	promService *service.PromV1Service,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			func(handler middleware.Handler) middleware.Handler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					ctx, span := otel.Tracer("grpc-middleware").Start(ctx, "middleware")
					defer span.End()
					trace.ContextWithSpanContext(ctx, span.SpanContext())
					return handler(ctx, req)
				}
			},
			recovery.Recovery(),
			logging.Server(logger),
			tracing.Server(tracing.WithTracerProvider(tp), tracing.WithTracerName("grpc")),
			ratelimit.Server(),
			metrics.Server(
				metrics.WithSeconds(prometheus.NewHistogram(prom.MetricSeconds)),
				metrics.WithRequests(prometheus.NewCounter(prom.MetricRequests)),
			),
			middler.IpMetric(prom.IpMetricCounter),
			validate.Validator(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	ping.RegisterPingServer(srv, pingService)
	promV1.RegisterPromServer(srv, promService)

	log.NewHelper(log.With(logger, "module", "server/grpc")).Info("grpc server initialized")

	return srv
}
