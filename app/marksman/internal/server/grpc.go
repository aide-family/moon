package server

import (
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

	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/service"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *conf.Bootstrap, namespaceService *service.NamespaceService, helper *klog.Helper) *grpc.Server {
	return newGRPCServer(bc.GetServer().GetGrpc(), bc.GetJwt(), namespaceService, helper)
}

func newGRPCServer(grpcConf conf.ServerConfig, jwtConf conf.JWTConfig, namespaceService *service.NamespaceService, helper *klog.Helper) *grpc.Server {
	selectorNamespaceMiddlewares := []middleware.Middleware{
		middler.MustNamespace(),
		middler.MustNamespaceExist(namespaceService.HasNamespace),
	}
	namespaceMiddleware := selector.Server(selectorNamespaceMiddlewares...).Match(middler.AllowListMatcher(namespaceAllowList...)).Build()
	selectorMustAuthMiddlewares := []middleware.Middleware{
		middler.JwtServe(jwtConf.GetSecret(), &jwt.JwtClaims{}),
		middler.MustLogin(),
		middler.BindJwtToken(),
		namespaceMiddleware,
	}
	authMiddleware := selector.Server(selectorMustAuthMiddlewares...).Match(middler.AllowListMatcher(authAllowList...)).Build()

	grpcMiddlewares := []middleware.Middleware{
		recovery.Recovery(),
		logging.Server(helper.Logger()),
		tracing.Server(),
		metadata.Server(),
		authMiddleware,
		middler.Validate(),
	}
	opts := []grpc.ServerOption{
		grpc.Middleware(grpcMiddlewares...),
	}
	if network := grpcConf.GetNetwork(); network != "" {
		opts = append(opts, grpc.Network(network))
	}
	if address := grpcConf.GetAddress(); address != "" {
		opts = append(opts, grpc.Address(address))
	}
	if timeout := grpcConf.GetTimeout(); timeout != nil {
		opts = append(opts, grpc.Timeout(timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	return srv
}
