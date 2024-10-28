package server

import (
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/plugin/slog"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *houyiconf.Bootstrap) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(recovery.WithHandler(slog.RecoveryHandle)),
			middleware.Logging(log.GetLogger()),
			middleware.Validate(protovalidate.WithFailFast(false)),
		),
	}
	grpcConf := bc.GetGrpc()
	if !types.IsNil(grpcConf) {
		if grpcConf.GetNetwork() != "" {
			opts = append(opts, grpc.Network(grpcConf.GetNetwork()))
		}
		if grpcConf.GetAddr() != "" {
			opts = append(opts, grpc.Address(grpcConf.GetAddr()))
		}
		if grpcConf.GetTimeout() != nil {
			opts = append(opts, grpc.Timeout(grpcConf.GetTimeout().AsDuration()))
		}
	}
	srv := grpc.NewServer(opts...)

	return srv
}
