package server

import (
	"embed"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	middle "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/cmd/palace/internal/service"
	"github.com/aide-family/moon/pkg/i18n"
	"github.com/aide-family/moon/pkg/metric"
	"github.com/aide-family/moon/pkg/middler"
	"github.com/aide-family/moon/pkg/util/docs"
)

//go:embed swagger
var docFS embed.FS

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	bc *conf.Bootstrap,
	healthService *service.HealthService,
	authService *service.AuthService,
	teamDatasourceService *service.TeamDatasourceService,
	menuService *service.MenuService,
	logger log.Logger,
) *http.Server {
	serverConf := bc.GetServer()
	httpConf := serverConf.GetHttp()
	jwtConf := bc.GetAuth().GetJwt()

	selectorMiddleware := []middle.Middleware{
		middleware.JwtServer(jwtConf.GetSignKey()),
		middleware.BindHeaders(menuService.GetMenuByOperation),
		middleware.MustLogin(authService.VerifyToken),
		middleware.MustPermission(authService.VerifyPermission),
	}
	authMiddleware := selector.Server(selectorMiddleware...).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()
	opts := []http.ServerOption{
		http.Filter(middler.Cors(httpConf)),
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			i18n.I18n(),
			logging.Server(logger),
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

	docs.RegisterDocs(srv, docFS, bc.IsDev())
	metric.RegisterRoutes(srv)
	registerOAuth2Routes(bc.GetAuth().GetOauth2(), srv, authService)
	registerTeamDatasourceRoutes(srv, teamDatasourceService)

	return srv
}

type OAuthService interface {
	OAuthLogin(app vobj.OAuthAPP) http.HandlerFunc
	OAuthLoginCallback(app vobj.OAuthAPP) http.HandlerFunc
}

func registerOAuth2Routes(c *conf.Auth_OAuth2, httpSrv *http.Server, authService OAuthService) {
	if !c.GetEnable() {
		return
	}
	auth := httpSrv.Route("/auth")
	list := c.GetConfigs()
	for _, config := range list {
		app := vobj.OAuthAPP(config.GetApp())
		appRoute := auth.Group(strings.ToLower(app.String()))
		appRoute.GET("/", authService.OAuthLogin(app))
		appRoute.GET("/callback", authService.OAuthLoginCallback(app))
	}
}

func registerTeamDatasourceRoutes(srv *http.Server, teamDatasourceService *service.TeamDatasourceService) {
	metricRoute := srv.Route("/api/team/datasource/metric")
	publicRoute := "/{datasourceId}/{target:[^/]+(?:/[^?]*)}"
	metricRoute.GET(publicRoute, teamDatasourceService.MetricDatasourceProxyHandler)
	metricRoute.POST(publicRoute, teamDatasourceService.MetricDatasourceProxyHandler)
	metricRoute.DELETE(publicRoute, teamDatasourceService.MetricDatasourceProxyHandler)
	metricRoute.PUT(publicRoute, teamDatasourceService.MetricDatasourceProxyHandler)
}
