package server

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/log"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *palaceconf.Bootstrap) *grpc.Server {
	c := bc.GetGrpc()
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(recovery.WithHandler(log.RecoveryHandle)),
			tracing.Server(),
			middleware.Logging(log.GetLogger()),
			middleware.Validate(protovalidate.WithFailFast(false)),
		),
	}
	if c.GetNetwork() != "" {
		opts = append(opts, grpc.Network(c.GetNetwork()))
	}
	if c.GetAddr() != "" {
		opts = append(opts, grpc.Address(c.GetAddr()))
	}
	if c.GetTimeout() != nil {
		opts = append(opts, grpc.Timeout(c.GetTimeout().AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	return srv
}
