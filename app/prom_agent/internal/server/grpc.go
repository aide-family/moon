package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"prometheus-manager/api/ping"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/service"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, pingService *service.PingService, logger log.Logger) *grpc.Server {
	logHelper := log.NewHelper(log.With(logger, "module", "server/grpc"))
	defer logHelper.Info("NewGRPCServer done")
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
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

	return srv
}
