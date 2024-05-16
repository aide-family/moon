package server

import (
	hookapi "github.com/aide-cloud/moon/api/rabbit/hook"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	v1 "github.com/aide-cloud/moon/api/helloworld/v1"
	pushapi "github.com/aide-cloud/moon/api/rabbit/push"
	"github.com/aide-cloud/moon/cmd/server/rabbit/internal/service"
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
	greeterService *service.GreeterService,
	configService *service.ConfigService,
	hookService *service.HookService,
) *Server {
	// 注册GRPC服务
	v1.RegisterGreeterServer(rpcSrv, greeterService)
	pushapi.RegisterConfigServer(rpcSrv, configService)
	hookapi.RegisterHookServer(rpcSrv, hookService)

	// 注册HTTP服务
	v1.RegisterGreeterHTTPServer(httpSrv, greeterService)
	pushapi.RegisterConfigHTTPServer(httpSrv, configService)
	r := httpSrv.Route("/")
	r.POST("/v1/hook/send/{route}", hookService.HookSendMsgHTTPHandler())

	return &Server{
		rpcSrv:  rpcSrv,
		httpSrv: httpSrv,
	}
}
