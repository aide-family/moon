package server

import (
	"github.com/aide-family/moon/api"
	alarmapi "github.com/aide-family/moon/api/admin/alarm"
	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	hookapi "github.com/aide-family/moon/api/admin/hook"
	inviteapi "github.com/aide-family/moon/api/admin/invite"
	menuapi "github.com/aide-family/moon/api/admin/menu"
	realtimeapi "github.com/aide-family/moon/api/admin/realtime"
	resourceapi "github.com/aide-family/moon/api/admin/resource"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	subscriberapi "github.com/aide-family/moon/api/admin/subscriber"
	teamapi "github.com/aide-family/moon/api/admin/team"
	userapi "github.com/aide-family/moon/api/admin/user"
	v1 "github.com/aide-family/moon/api/helloworld/v1"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/alarm"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/dict"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/invite"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/menu"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/subscriber"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/team"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/user"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, RegisterService)

// Server 服务
type Server struct {
	rpcSrv        *grpc.Server
	httpSrv       *http.Server
	strategyWatch *StrategyWatch
}

// GetRPCServer 获取rpc server
func (s *Server) GetRPCServer() *grpc.Server {
	return s.rpcSrv
}

// GetHTTPServer 获取http server
func (s *Server) GetHTTPServer() *http.Server {
	return s.httpSrv
}

// GetServers 注册服务
func (s *Server) GetServers() []transport.Server {
	return []transport.Server{
		s.rpcSrv,
		s.httpSrv,
		s.strategyWatch,
	}
}

// RegisterService 注册服务
func RegisterService(
	c *palaceconf.Bootstrap,
	data *data.Data,
	alertService *service.AlertService,
	serverService *service.ServerService,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	greeter *service.GreeterService,
	healthService *service.HealthService,
	userService *user.Service,
	authorizationService *authorization.Service,
	resourceService *resource.Service,
	teamService *team.Service,
	teamRoleService *team.RoleService,
	datasourceService *datasource.Service,
	menuService *menu.Service,
	metricService *datasource.MetricService,
	dictService *dict.Service,
	strategyService *strategy.Service,
	strategyTemplateService *strategy.TemplateService,
	dashboardService *realtime.DashboardService,
	alarmService *realtime.AlarmService,
	alarmPageSelfService *realtime.AlarmPageSelfService,
	alarmGroupService *alarm.GroupService,
	subscriberService *subscriber.Service,
	hookService *hook.Service,
	inviteService *invite.Service,
	messageService *user.MessageService,
) *Server {
	// 注册GRPC服务
	v1.RegisterGreeterServer(rpcSrv, greeter)
	userapi.RegisterUserServer(rpcSrv, userService)
	authorizationapi.RegisterAuthorizationServer(rpcSrv, authorizationService)
	resourceapi.RegisterResourceServer(rpcSrv, resourceService)
	menuapi.RegisterMenuServer(rpcSrv, menuService)
	teamapi.RegisterTeamServer(rpcSrv, teamService)
	teamapi.RegisterRoleServer(rpcSrv, teamRoleService)
	datasourceapi.RegisterDatasourceServer(rpcSrv, datasourceService)
	datasourceapi.RegisterMetricServer(rpcSrv, metricService)
	dictapi.RegisterDictServer(rpcSrv, dictService)
	api.RegisterHealthServer(rpcSrv, healthService)
	strategyapi.RegisterStrategyServer(rpcSrv, strategyService)
	strategyapi.RegisterTemplateServer(rpcSrv, strategyTemplateService)
	realtimeapi.RegisterDashboardServer(rpcSrv, dashboardService)
	realtimeapi.RegisterAlarmServer(rpcSrv, alarmService)
	realtimeapi.RegisterAlarmPageSelfServer(rpcSrv, alarmPageSelfService)
	alarmapi.RegisterAlarmServer(rpcSrv, alarmGroupService)
	subscriberapi.RegisterSubscriberServer(rpcSrv, subscriberService)
	hookapi.RegisterHookServer(rpcSrv, hookService)
	api.RegisterAlertServer(rpcSrv, alertService)
	userapi.RegisterMessageServer(rpcSrv, messageService)
	inviteapi.RegisterInviteServer(rpcSrv, inviteService)
	api.RegisterServerServer(rpcSrv, serverService)

	// 注册HTTP服务
	v1.RegisterGreeterHTTPServer(httpSrv, greeter)
	userapi.RegisterUserHTTPServer(httpSrv, userService)
	authorizationapi.RegisterAuthorizationHTTPServer(httpSrv, authorizationService)
	resourceapi.RegisterResourceHTTPServer(httpSrv, resourceService)
	menuapi.RegisterMenuHTTPServer(httpSrv, menuService)
	teamapi.RegisterTeamHTTPServer(httpSrv, teamService)
	teamapi.RegisterRoleHTTPServer(httpSrv, teamRoleService)
	datasourceapi.RegisterDatasourceHTTPServer(httpSrv, datasourceService)
	datasourceapi.RegisterMetricHTTPServer(httpSrv, metricService)
	dictapi.RegisterDictHTTPServer(httpSrv, dictService)
	api.RegisterHealthHTTPServer(httpSrv, healthService)
	strategyapi.RegisterStrategyHTTPServer(httpSrv, strategyService)
	strategyapi.RegisterTemplateHTTPServer(httpSrv, strategyTemplateService)
	realtimeapi.RegisterDashboardHTTPServer(httpSrv, dashboardService)
	realtimeapi.RegisterAlarmHTTPServer(httpSrv, alarmService)
	realtimeapi.RegisterAlarmPageSelfHTTPServer(httpSrv, alarmPageSelfService)
	alarmapi.RegisterAlarmHTTPServer(httpSrv, alarmGroupService)
	subscriberapi.RegisterSubscriberHTTPServer(httpSrv, subscriberService)
	hookapi.RegisterHookHTTPServer(httpSrv, hookService)
	api.RegisterAlertHTTPServer(httpSrv, alertService)
	inviteapi.RegisterInviteHTTPServer(httpSrv, inviteService)
	userapi.RegisterMessageHTTPServer(httpSrv, messageService)
	api.RegisterServerHTTPServer(httpSrv, serverService)

	// custom api
	proxy := httpSrv.Route("/v1")
	proxy.GET("/proxy", datasourceService.ProxyQuery)
	proxy.POST("/proxy", datasourceService.ProxyQuery)

	auth := httpSrv.Route("/auth")
	auth.GET("/github", authorizationService.OAuthLogin(vobj.OAuthAPPGithub))
	auth.GET("/github/callback", authorizationService.OAuthLoginCallback(vobj.OAuthAPPGithub))
	auth.GET("/gitee", authorizationService.OAuthLogin(vobj.OAuthAPPGitee))
	auth.GET("/gitee/callback", authorizationService.OAuthLoginCallback(vobj.OAuthAPPGitee))

	// 是否启动链路追踪
	if !types.IsNil(c.GetTracer()) {
		var err error
		tracerConf := c.GetTracer()
		switch tracerConf.GetDriver() {
		// TODO other tracer
		default:
			err = conn.InitJaegerTracer("moon.palace", tracerConf.GetJaeger().GetEndpoint())
		}
		if !types.IsNil(err) {
			panic(err)
		}
	}

	return &Server{
		rpcSrv:        rpcSrv,
		httpSrv:       httpSrv,
		strategyWatch: newStrategyWatch(c, data, alertService),
	}
}
