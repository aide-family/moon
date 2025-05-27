package server

import (
	"github.com/go-kratos/kratos/v2/log"
	middle "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/cmd/palace/internal/service"
	portal_service "github.com/aide-family/moon/cmd/palace/internal/service/portal"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/middler"
)

func NewPortalHTTPServer(
	bc *conf.Bootstrap,
	healthService *service.HealthService,
	menuService *service.MenuService,
	authService *portal_service.AuthService,
	logger log.Logger,
) *PortalHTTPServer {
	httpConf := bc.GetPortal()
	jwtConf := bc.GetAuth().GetJwt()

	selectorMiddleware := []middle.Middleware{
		middleware.JwtServer(jwtConf.GetSignKey()),
		middleware.MustLogin(authService.VerifyToken),
	}
	authMiddleware := selector.Server(selectorMiddleware...).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()
	opts := []http.ServerOption{
		http.Filter(middler.Cors(httpConf)),
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			merr.I18n(),
			logging.Server(logger),
			middleware.BindHeaders(menuService.GetMenuByOperation),
			authMiddleware,
			middler.Validate(),
			middleware.OperateLog(healthService.CreateOperateLog),
		),
	}
	if httpConf.GetNetwork() != "" {
		opts = append(opts, http.Network(httpConf.GetNetwork()))
	}
	if httpConf.GetAddr() != "" {
		opts = append(opts, http.Address(httpConf.GetAddr()))
	}
	if httpConf.GetTimeout() != nil {
		opts = append(opts, http.Timeout(httpConf.GetTimeout().AsDuration()))
	}
	srv := http.NewServer(opts...)

	registerOAuth2Routes(bc.GetAuth().GetOauth2Portal(), srv, authService)
	return &PortalHTTPServer{
		Server: srv,
	}
}

type PortalHTTPServer struct {
	*http.Server
}
