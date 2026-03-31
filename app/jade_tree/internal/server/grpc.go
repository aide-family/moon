package server

import (
	healthv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/jwt"
	"github.com/aide-family/magicbox/server/middler"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/aide-family/jade_tree/internal/conf"
	"github.com/aide-family/jade_tree/internal/service"
)

func NewGRPCServer(bc *conf.Bootstrap, helper *klog.Helper) *grpc.Server {
	return newGRPCServer(bc.GetServer().GetGrpc(), bc.GetJwt(), helper)
}

func newGRPCServer(grpcConf conf.ServerConfig, jwtConf conf.JWTConfig, helper *klog.Helper) *grpc.Server {
	authMiddleware := selector.Server(middler.JwtServe(jwtConf.GetSecret(), &jwt.JwtClaims{}), middler.MustLogin(), middler.BindJwtToken()).Match(middler.AllowListMatcher(authAllowList...)).Build()
	grpcMiddlewares := []middleware.Middleware{recovery.Recovery(), logging.Server(helper.Logger()), tracing.Server(), metadata.Server(), authMiddleware, middler.Validate()}
	opts := []grpc.ServerOption{grpc.Middleware(grpcMiddlewares...)}
	if network := grpcConf.GetNetwork(); network != "" {
		opts = append(opts, grpc.Network(network))
	}
	if address := grpcConf.GetAddress(); address != "" {
		opts = append(opts, grpc.Address(address))
	}
	if timeout := grpcConf.GetTimeout(); timeout != nil {
		opts = append(opts, grpc.Timeout(timeout.AsDuration()))
	}
	return grpc.NewServer(opts...)
}

func RegisterGRPCService(grpcSrv *grpc.Server, healthService *service.HealthService) Servers {
	healthv1.RegisterHealthServer(grpcSrv, healthService)
	return Servers{newServer("grpc", grpcSrv)}
}
