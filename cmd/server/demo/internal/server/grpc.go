package server

import (
	"github.com/aide-family/moon/cmd/server/demo/internal/democonf"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/plugin/mlog"
	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *democonf.Bootstrap) *grpc.Server {
	c := bc.GetServer()
	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(recovery.WithHandler(mlog.RecoveryHandle)),
			middleware.Validate(protovalidate.WithFailFast(true)),
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

	return srv
}
