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
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/service"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *conf.Bootstrap, namespaceService *service.NamespaceService, helper *klog.Helper) *http.Server {
	return newHTTPServer(bc.GetServer().GetHttp(), bc.GetJwt(), namespaceService, helper)
}

func newHTTPServer(httpConf conf.ServerConfig, jwtConf conf.JWTConfig, namespaceService *service.NamespaceService, helper *klog.Helper) *http.Server {
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

	httpMiddlewares := []middleware.Middleware{
		recovery.Recovery(),
		logging.Server(helper.Logger()),
		tracing.Server(),
		metadata.Server(),
		authMiddleware,
		middler.Validate(),
	}

	opts := []http.ServerOption{
		middler.DefaultCors(),
		http.Middleware(httpMiddlewares...),
	}
	if network := httpConf.GetNetwork(); network != "" {
		opts = append(opts, http.Network(network))
	}
	if address := httpConf.GetAddress(); address != "" {
		opts = append(opts, http.Address(address))
	}
	if timeout := httpConf.GetTimeout(); timeout != nil {
		opts = append(opts, http.Timeout(timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	return srv
}
