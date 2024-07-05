package server

import (
	"github.com/aide-family/moon/api"
	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	menuapi "github.com/aide-family/moon/api/admin/menu"
	resourceapi "github.com/aide-family/moon/api/admin/resource"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	teamapi "github.com/aide-family/moon/api/admin/team"
	userapi "github.com/aide-family/moon/api/admin/user"
	v1 "github.com/aide-family/moon/api/helloworld/v1"
	"github.com/aide-family/moon/cmd/server/palace/internal/service"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/dict"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/menu"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/team"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/user"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, RegisterService)

type Server struct {
	rpcSrv  *grpc.Server
	httpSrv *http.Server
}

// GetRpcServer 获取rpc server
func (s *Server) GetRpcServer() *grpc.Server {
	return s.rpcSrv
}

// GetHttpServer 获取http server
func (s *Server) GetHttpServer() *http.Server {
	return s.httpSrv
}

// GetServers 注册服务
func (s *Server) GetServers() []transport.Server {
	return []transport.Server{
		s.rpcSrv,
		s.httpSrv,
	}
}

func RegisterService(
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
	menuService *menu.MenuService,
	metricService *datasource.MetricService,
	dictService *dict.Service,
	strategyService *strategy.Service,
	strategyTemplateService *strategy.TemplateService,
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

	return &Server{
		rpcSrv:  rpcSrv,
		httpSrv: httpSrv,
	}
}
