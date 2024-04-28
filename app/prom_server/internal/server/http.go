package server

import (
	"context"
	nHttp "net/http"

	"github.com/aide-family/moon/api/alarm/hook"
	"github.com/aide-family/moon/api/interflows"
	"github.com/aide-family/moon/api/ping"
	"github.com/aide-family/moon/api/server/alarm/history"
	"github.com/aide-family/moon/api/server/alarm/realtime"
	"github.com/aide-family/moon/api/server/auth"
	"github.com/aide-family/moon/api/server/dashboard"
	"github.com/aide-family/moon/api/server/prom/endpoint"
	"github.com/aide-family/moon/api/server/prom/notify"
	"github.com/aide-family/moon/api/server/prom/strategy"
	"github.com/aide-family/moon/api/server/prom/strategy/group"
	"github.com/aide-family/moon/api/server/system"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/conf"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/app/prom_server/internal/service"
	"github.com/aide-family/moon/app/prom_server/internal/service/alarmservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/authservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/dashboardservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/interflowservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/promservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/systemservice"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/helper/prom"
	"github.com/aide-family/moon/pkg/servers"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpServer struct {
	*http.Server
}

// RegisterHttpServer new a HTTP server register.
func RegisterHttpServer(
	srv *http.Server,
	pingService *service.PingService,
	dictService *systemservice.Service,
	strategyService *promservice.StrategyService,
	strategyGroupService *promservice.GroupService,
	hookService *alarmservice.HookService,
	historyService *alarmservice.HistoryService,
	authService *authservice.AuthService,
	userService *systemservice.UserService,
	roleService *systemservice.RoleService,
	endpointService *promservice.EndpointService,
	apiService *systemservice.ApiService,
	chatGroupService *promservice.ChatGroupService,
	notifyService *promservice.NotifyService,
	realtimeService *alarmservice.RealtimeService,
	interflowService *interflowservice.HookInterflowService,
	chartService *dashboardservice.ChartService,
	dashboardService *dashboardservice.DashboardService,
	syslogService *systemservice.SyslogService,
	templateService *promservice.TemplateService,
) *HttpServer {
	ping.RegisterPingHTTPServer(srv, pingService)
	system.RegisterDictHTTPServer(srv, dictService)
	strategy.RegisterStrategyHTTPServer(srv, strategyService)
	group.RegisterGroupHTTPServer(srv, strategyGroupService)
	hook.RegisterHookHTTPServer(srv, hookService)
	history.RegisterHistoryHTTPServer(srv, historyService)
	auth.RegisterAuthHTTPServer(srv, authService)
	system.RegisterUserHTTPServer(srv, userService)
	system.RegisterRoleHTTPServer(srv, roleService)
	system.RegisterSyslogHTTPServer(srv, syslogService)
	endpoint.RegisterEndpointHTTPServer(srv, endpointService)
	system.RegisterApiHTTPServer(srv, apiService)
	notify.RegisterNotifyHTTPServer(srv, notifyService)
	notify.RegisterChatGroupHTTPServer(srv, chatGroupService)
	realtime.RegisterRealtimeHTTPServer(srv, realtimeService)
	interflows.RegisterHookInterflowHTTPServer(srv, interflowService)
	dashboard.RegisterDashboardHTTPServer(srv, dashboardService)
	dashboard.RegisterChartHTTPServer(srv, chartService)
	notify.RegisterTemplateHTTPServer(srv, templateService)

	srv.Route("/api").POST("/upload", func(ctx http.Context) error {
		return ctx.Result(nHttp.StatusOK, "ok")
	})

	return &HttpServer{Server: srv}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	jwtConf *conf.JWT,
	d *data.Data,
	apiWhite *conf.ApiWhite,
	logger log.Logger,
) *http.Server {
	logHelper := log.NewHelper(log.With(logger, "module", "http"))
	defer logHelper.Info("NewHTTPServer done")

	jwt.WithSigningMethod(jwtv4.SigningMethodHS256)
	jwt.WithClaims(func() jwtv4.Claims { return &jwtv4.RegisteredClaims{} })
	middler.SetSecret(jwtConf.GetSecret())
	middler.SetExpire(jwtConf.GetExpires().AsDuration())
	middler.SetIssuer(jwtConf.GetIssuer())

	allApi := apiWhite.GetAll()
	jwtApis := append(allApi, apiWhite.GetJwtApi()...)
	rbacApis := append(allApi, apiWhite.GetRbacApi()...)

	jwtMiddle := selector.Server(
		middler.JwtServer(),
		middler.MustLogin(d.Cache()),
	).Match(middler.NewWhiteListMatcher(jwtApis)).Build()
	rbacMiddle := selector.Server(middler.RbacServer(
		func(ctx context.Context, userID, roleID uint32) error {
			return do.CheckUserRoleExist(ctx, d.Cache(), userID, roleID)
		},
		func(ctx context.Context, path, method string) (uint64, error) {
			return do.GetApiIDByPathAndMethod(d.Cache(), path, method)
		},
	)).Match(middler.NewWhiteListMatcher(rbacApis)).Build()

	var opts = []http.ServerOption{
		http.Filter(middler.Cors(), middler.Context(), middler.LocalHttpRequestFilter()),
		http.Middleware(
			middler.IpMetric(prom.IpMetricCounter),
			recovery.Recovery(),
			middler.Logging(logger),
			jwtMiddle,
			rbacMiddle,
			validate.Validator(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	srv.HandlePrefix("/msg", nHttp.HandlerFunc(func(w nHttp.ResponseWriter, r *nHttp.Request) {
		sendCh <- &servers.Message{
			MsgType: 1,
			Content: "你有新的告警了",
			Title:   "告警通知",
			Biz:     "alarm",
		}
		_, _ = w.Write([]byte("ok"))
	}))
	srv.HandlePrefix("/metrics", promhttp.Handler())
	// doc
	srv.HandlePrefix("/doc/", nHttp.StripPrefix("/doc/", nHttp.FileServer(nHttp.Dir("../../third_party/swagger_ui"))))
	return srv
}
