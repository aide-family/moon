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
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/jade_tree/internal/conf"
	"github.com/aide-family/jade_tree/internal/service"
)

func NewHTTPServer(bc *conf.Bootstrap, helper *klog.Helper) *http.Server {
	return newHTTPServer(bc.GetServer().GetHttp(), bc.GetJwt(), helper)
}

func newHTTPServer(httpConf conf.ServerConfig, jwtConf conf.JWTConfig, helper *klog.Helper) *http.Server {
	authMiddleware := selector.Server(middler.JwtServe(jwtConf.GetSecret(), &jwt.JwtClaims{}), middler.MustLogin(), middler.BindJwtToken()).Match(middler.AllowListMatcher(authAllowList...)).Build()
	httpMiddlewares := []middleware.Middleware{recovery.Recovery(), logging.Server(helper.Logger()), tracing.Server(), metadata.Server(), authMiddleware, middler.Validate()}
	opts := []http.ServerOption{middler.DefaultCors(), http.Middleware(httpMiddlewares...)}
	if network := httpConf.GetNetwork(); network != "" {
		opts = append(opts, http.Network(network))
	}
	if address := httpConf.GetAddress(); address != "" {
		opts = append(opts, http.Address(address))
	}
	if timeout := httpConf.GetTimeout(); timeout != nil {
		opts = append(opts, http.Timeout(timeout.AsDuration()))
	}
	return http.NewServer(opts...)
}

func RegisterHTTPService(httpSrv *http.Server, healthService *service.HealthService) Servers {
	healthv1.RegisterHealthHTTPServer(httpSrv, healthService)
	return Servers{newServer("http", httpSrv)}
}

var authAllowList = []string{healthv1.OperationHealthHealthCheck}
