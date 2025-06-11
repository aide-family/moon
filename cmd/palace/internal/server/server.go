package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/moon-monitor/moon/cmd/palace/internal/conf"
	"github.com/moon-monitor/moon/cmd/palace/internal/service"
	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	NewTickerServer,
	RegisterService,
)

// RegisterService register service
func RegisterService(
	c *conf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	tickerSrv *TickerServer,
	healthService *service.HealthService,
	authService *service.AuthService,
	serverService *service.ServerService,
	resourceService *service.ResourceService,
	userService *service.UserService,
	callbackService *service.CallbackService,
	teamDashboardService *service.TeamDashboardService,
	datasourceService *service.TeamDatasourceService,
	dictService *service.TeamDictService,
	noticeService *service.TeamNoticeService,
	strategyService *service.TeamStrategyService,
	teamService *service.TeamService,
	systemService *service.SystemService,
	teamLogService *service.TeamLogService,
	alertService *service.AlertService,
) server.Servers {
	common.RegisterHealthServer(rpcSrv, healthService)
	common.RegisterServerServer(rpcSrv, serverService)
	palace.RegisterAlertServer(rpcSrv, alertService)

	common.RegisterHealthHTTPServer(httpSrv, healthService)
	common.RegisterServerHTTPServer(httpSrv, serverService)
	palace.RegisterAuthHTTPServer(httpSrv, authService)
	palace.RegisterResourceHTTPServer(httpSrv, resourceService)
	palace.RegisterUserHTTPServer(httpSrv, userService)
	palace.RegisterCallbackHTTPServer(httpSrv, callbackService)
	palace.RegisterTeamDashboardHTTPServer(httpSrv, teamDashboardService)
	palace.RegisterTeamDatasourceHTTPServer(httpSrv, datasourceService)
	palace.RegisterTeamDictHTTPServer(httpSrv, dictService)
	palace.RegisterTeamNoticeHTTPServer(httpSrv, noticeService)
	palace.RegisterTeamStrategyHTTPServer(httpSrv, strategyService)
	palace.RegisterTeamHTTPServer(httpSrv, teamService)
	palace.RegisterSystemHTTPServer(httpSrv, systemService)
	palace.RegisterTeamLogHTTPServer(httpSrv, teamLogService)
	palace.RegisterAlertHTTPServer(httpSrv, alertService)
	return server.Servers{rpcSrv, httpSrv, tickerSrv}
}
