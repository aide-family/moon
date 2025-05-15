package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/cmd/laurel/internal/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/middler"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *conf.Bootstrap, logger log.Logger) *grpc.Server {
	serverConf := bc.GetServer()
	grpcConf := serverConf.GetGrpc()
	jwtConf := bc.GetAuth().GetJwt()

	authMiddleware := selector.Server(
		middleware.JwtServer(jwtConf.GetSignKey()),
	).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()

	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			merr.I18n(),
			logging.Server(logger),
			authMiddleware,
			middler.Validate(),
		),
	}
	if grpcConf.GetNetwork() != "" {
		opts = append(opts, grpc.Network(grpcConf.GetNetwork()))
	}
	if grpcConf.GetAddr() != "" {
		opts = append(opts, grpc.Address(grpcConf.GetAddr()))
	}
	if grpcConf.GetTimeout() != nil {
		opts = append(opts, grpc.Timeout(grpcConf.GetTimeout().AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	return srv
}
