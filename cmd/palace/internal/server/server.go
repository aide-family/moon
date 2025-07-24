// Package server is a server package for kratos.
package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/service"
	portal_service "github.com/aide-family/moon/cmd/palace/internal/service/portal"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	portalapi "github.com/aide-family/moon/pkg/api/palace/portal"
	"github.com/aide-family/moon/pkg/plugin/server"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	NewPortalHTTPServer,
	NewTickerServer,
	NewEvent,
	RegisterService,
)

// RegisterService register service
func RegisterService(
	c *conf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	portalHTTPSrv *PortalHTTPServer,
	tickerSrv *TickerServer,
	eventSrv *Event,
	healthService *service.HealthService,
	authService *service.AuthService,
	serverService *service.ServerService,
	menuService *service.MenuService,
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
	timeEngineService *service.TimeEngineService,
	portalAuthService *portal_service.AuthService,
	portalHomeService *portal_service.HomeService,
	portalPricingService *portal_service.PricingService,
	teamStrategyMetricService *service.TeamStrategyMetricService,
) server.Servers {
	common.RegisterHealthServer(rpcSrv, healthService)
	common.RegisterServerServer(rpcSrv, serverService)
	palace.RegisterAlertServer(rpcSrv, alertService)
	palace.RegisterCallbackServer(rpcSrv, callbackService)

	common.RegisterHealthHTTPServer(httpSrv, healthService)
	common.RegisterServerHTTPServer(httpSrv, serverService)
	palace.RegisterAuthHTTPServer(httpSrv, authService)
	palace.RegisterMenuHTTPServer(httpSrv, menuService)
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
	palace.RegisterTimeEngineHTTPServer(httpSrv, timeEngineService)
	palace.RegisterTeamStrategyMetricHTTPServer(httpSrv, teamStrategyMetricService)

	// portal
	portalapi.RegisterAuthHTTPServer(portalHTTPSrv.Server, portalAuthService)
	portalapi.RegisterHomeHTTPServer(portalHTTPSrv.Server, portalHomeService)
	portalapi.RegisterPricingHTTPServer(portalHTTPSrv.Server, portalPricingService)

	return server.Servers{rpcSrv, httpSrv, portalHTTPSrv, tickerSrv, eventSrv}
}
