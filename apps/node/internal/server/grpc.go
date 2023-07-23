package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	loadV1 "prometheus-manager/api/strategy/v1/load"
	pullV1 "prometheus-manager/api/strategy/v1/pull"
	pushV1 "prometheus-manager/api/strategy/v1/push"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	pushService *service.PushService,
	loadService *service.LoadService,
	pullService *service.PullService,
	logger log.Logger,
) *grpc.Server {
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
	pushV1.RegisterPushServer(srv, pushService)
	pullV1.RegisterPullServer(srv, pullService)
	loadV1.RegisterLoadServer(srv, loadService)

	log.NewHelper(log.With(logger, "module", "server/grpc")).Info("grpc server initialized")

	return srv
}
