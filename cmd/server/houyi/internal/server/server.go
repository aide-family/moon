package server

import (
	"github.com/aide-cloud/moon/api"
	v1 "github.com/aide-cloud/moon/api/helloworld/v1"
	metadataapi "github.com/aide-cloud/moon/api/houyi/metadata"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/service"

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
	metricService *service.MetricService,
	healthService *service.HealthService,
) *Server {
	// 注册GRPC服务
	v1.RegisterGreeterServer(rpcSrv, greeter)
	metadataapi.RegisterMetricServer(rpcSrv, metricService)
	api.RegisterHealthServer(rpcSrv, healthService)
	// 注册HTTP服务
	v1.RegisterGreeterHTTPServer(httpSrv, greeter)
	metadataapi.RegisterMetricHTTPServer(httpSrv, metricService)
	api.RegisterHealthHTTPServer(httpSrv, healthService)

	return &Server{
		rpcSrv:  rpcSrv,
		httpSrv: httpSrv,
	}
}
