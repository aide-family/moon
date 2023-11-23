package server

import (
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"prometheus-manager/api/alarm/history"
	"prometheus-manager/api/alarm/hook"
	"prometheus-manager/api/alarm/page"
	"prometheus-manager/api/auth"
	"prometheus-manager/api/dict"
	"prometheus-manager/api/ping"
	"prometheus-manager/api/prom/strategy"
	"prometheus-manager/api/prom/strategy/group"
	"prometheus-manager/api/system"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/service"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
	"prometheus-manager/app/prom_server/internal/service/authservice"
	"prometheus-manager/app/prom_server/internal/service/dictservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"
	"prometheus-manager/app/prom_server/internal/service/systemservice"
	"prometheus-manager/pkg/helper"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

type HttpServer struct {
	*http.Server
}

// RegisterHttpServer new a HTTP server register.
func RegisterHttpServer(
	srv *http.Server,
	pingService *service.PingService,
	dictService *dictservice.Service,
	strategyService *promservice.StrategyService,
	strategyGroupService *promservice.GroupService,
	alarmPageService *alarmservice.AlarmPageService,
	hookService *alarmservice.HookService,
	historyService *alarmservice.HistoryService,
	authService *authservice.AuthService,
	userService *systemservice.UserService,
	roleService *systemservice.RoleService,
) *HttpServer {
	ping.RegisterPingHTTPServer(srv, pingService)
	dict.RegisterDictHTTPServer(srv, dictService)
	strategy.RegisterStrategyHTTPServer(srv, strategyService)
	group.RegisterGroupHTTPServer(srv, strategyGroupService)
	page.RegisterAlarmPageHTTPServer(srv, alarmPageService)
	hook.RegisterHookHTTPServer(srv, hookService)
	history.RegisterHistoryHTTPServer(srv, historyService)
	auth.RegisterAuthHTTPServer(srv, authService)
	system.RegisterUserHTTPServer(srv, userService)
	system.RegisterRoleHTTPServer(srv, roleService)

	return &HttpServer{Server: srv}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	whiteList *conf.WhiteList,
	logger log.Logger,
) *http.Server {
	logHelper := log.NewHelper(log.With(logger, "module", "http"))
	defer logHelper.Info("NewHTTPServer done")

	jwt.WithSigningMethod(jwtv4.SigningMethodHS256)
	jwt.WithClaims(func() jwtv4.Claims { return &jwtv4.RegisteredClaims{} })

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			selector.Server(helper.JwtServer(), validate.Validator()).Match(helper.NewWhiteListMatcher(whiteList.GetApi())).Build(),
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
	srv.HandlePrefix("/metrics", promhttp.Handler())

	return srv
}
