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
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/metric"
	"github.com/aide-family/moon/pkg/middler"
	"github.com/aide-family/moon/pkg/util/docs"
)

//go:embed swagger
var docFS embed.FS

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	bc *conf.Bootstrap,
	authService *service.AuthService,
	teamDatasourceService *service.TeamDatasourceService,
	logger log.Logger,
) *http.Server {
	serverConf := bc.GetServer()
	httpConf := serverConf.GetHttp()
	jwtConf := bc.GetAuth().GetJwt()

	selectorMiddleware := []middle.Middleware{
		middleware.JwtServer(jwtConf.GetSignKey()),
		middleware.BindHeaders(),
		middleware.MustLogin(authService.VerifyToken),
		middleware.MustPermission(authService.VerifyPermission),
	}
	authMiddleware := selector.Server(selectorMiddleware...).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()
	opts := []http.ServerOption{
		http.Filter(middler.Cors(httpConf)),
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			merr.I18n(),
			logging.Server(logger),
			authMiddleware,
			middler.Validate(),
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

func registerOAuth2Routes(c *conf.Auth_OAuth2, httpSrv *http.Server, authService *service.AuthService) {
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
	metricRoute := srv.Route("/datasource/metric")
	publicRoute := "/{datasourceId}/{target:[^/]+(?:/[^?]*)}"
	metricRoute.GET(publicRoute, teamDatasourceService.MetricDatasourceProxy)
	metricRoute.POST(publicRoute, teamDatasourceService.MetricDatasourceProxy)
	metricRoute.DELETE(publicRoute, teamDatasourceService.MetricDatasourceProxy)
	metricRoute.PUT(publicRoute, teamDatasourceService.MetricDatasourceProxy)
}
