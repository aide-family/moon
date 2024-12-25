package server

import (
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/plugin/mlog"
	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *rabbitconf.Bootstrap) *grpc.Server {
	c := bc.GetGrpc()
	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(recovery.WithHandler(mlog.RecoveryHandle)),
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
