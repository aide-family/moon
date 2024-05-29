package server

import (
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	authorizationapi "github.com/aide-cloud/moon/api/admin/authorization"
	datasourceapi "github.com/aide-cloud/moon/api/admin/datasource"
	resourceapi "github.com/aide-cloud/moon/api/admin/resource"
	teamapi "github.com/aide-cloud/moon/api/admin/team"
	userapi "github.com/aide-cloud/moon/api/admin/user"
	v1 "github.com/aide-cloud/moon/api/helloworld/v1"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/datasource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/resource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/team"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/user"
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
	userService *user.Service,
	authorizationService *authorization.Service,
	resourceService *resource.Service,
	teamService *team.Service,
	teamRoleService *team.RoleService,
	datasourceService *datasource.Service,
	teamMenuService *resource.MenuService,
) *Server {
	// 注册GRPC服务
	v1.RegisterGreeterServer(rpcSrv, greeter)
	userapi.RegisterUserServer(rpcSrv, userService)
	authorizationapi.RegisterAuthorizationServer(rpcSrv, authorizationService)
	resourceapi.RegisterResourceServer(rpcSrv, resourceService)
	resourceapi.RegisterMenuServer(rpcSrv, teamMenuService)
	teamapi.RegisterTeamServer(rpcSrv, teamService)
	teamapi.RegisterRoleServer(rpcSrv, teamRoleService)
	datasourceapi.RegisterDatasourceServer(rpcSrv, datasourceService)

	// 注册HTTP服务
	v1.RegisterGreeterHTTPServer(httpSrv, greeter)
	userapi.RegisterUserHTTPServer(httpSrv, userService)
	authorizationapi.RegisterAuthorizationHTTPServer(httpSrv, authorizationService)
	resourceapi.RegisterResourceHTTPServer(httpSrv, resourceService)
	resourceapi.RegisterMenuHTTPServer(httpSrv, teamMenuService)
	teamapi.RegisterTeamHTTPServer(httpSrv, teamService)
	teamapi.RegisterRoleHTTPServer(httpSrv, teamRoleService)
	datasourceapi.RegisterDatasourceHTTPServer(httpSrv, datasourceService)

	return &Server{
		rpcSrv:  rpcSrv,
		httpSrv: httpSrv,
	}
}
