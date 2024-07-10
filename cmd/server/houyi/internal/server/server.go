package server

import (
	"github.com/aide-family/moon/api"
	v1 "github.com/aide-family/moon/api/helloworld/v1"
	metadataapi "github.com/aide-family/moon/api/houyi/metadata"
	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, RegisterService)

type Server struct {
	rpcSrv        *grpc.Server
	httpSrv       *http.Server
	strategyWatch *watch.Watcher
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
		s.strategyWatch,
	}
}

func RegisterService(
	c *houyiconf.Bootstrap,
	data *data.Data,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	greeter *service.GreeterService,
	metricService *service.MetricService,
	healthService *service.HealthService,
	strategyService *service.StrategyService,
) *Server {
	// 注册GRPC服务
	v1.RegisterGreeterServer(rpcSrv, greeter)
	metadataapi.RegisterMetricServer(rpcSrv, metricService)
	api.RegisterHealthServer(rpcSrv, healthService)
	strategyapi.RegisterStrategyServer(rpcSrv, strategyService)
	// 注册HTTP服务
	v1.RegisterGreeterHTTPServer(httpSrv, greeter)
	metadataapi.RegisterMetricHTTPServer(httpSrv, metricService)
	api.RegisterHealthHTTPServer(httpSrv, healthService)
	strategyapi.RegisterStrategyHTTPServer(httpSrv, strategyService)

	return &Server{
		rpcSrv:        rpcSrv,
		httpSrv:       httpSrv,
		strategyWatch: newStrategyWatch(c, data),
	}
}
