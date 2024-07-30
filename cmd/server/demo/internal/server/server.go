package server

import (
	v1 "github.com/aide-family/moon/api/helloworld/v1"
	v1api "github.com/aide-family/moon/api/v1"
	"github.com/aide-family/moon/cmd/server/demo/internal/service"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, RegisterService)

// Server moon server register providers.
type Server struct {
	rpcSrv  *grpc.Server
	httpSrv *http.Server
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
	}
}

// RegisterService register moon service
func RegisterService(
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	greeter *service.GreeterService,
	helloService *service.HelloService,
) *Server {
	// 注册GRPC服务
	v1.RegisterGreeterServer(rpcSrv, greeter)
	v1api.RegisterHelloServer(rpcSrv, helloService)

	// 注册HTTP服务
	v1.RegisterGreeterHTTPServer(httpSrv, greeter)
	v1api.RegisterHelloHTTPServer(httpSrv, helloService)

	return &Server{
		rpcSrv:  rpcSrv,
		httpSrv: httpSrv,
	}
}
