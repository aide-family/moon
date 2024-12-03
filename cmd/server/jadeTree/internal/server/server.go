package server

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/data"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/jadetreeconf"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/service"
	"github.com/aide-family/moon/pkg/watch"

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
	heartbeatSrv  *HeartbeatServer
	consumerWatch *watch.Watcher
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
		s.heartbeatSrv,
		s.consumerWatch,
	}
}

// RegisterService 注册服务
func RegisterService(
	bc *jadetreeconf.Bootstrap,
	data *data.Data,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	healthService *service.HealthService,
) *Server {
	// 注册GRPC服务
	api.RegisterHealthServer(rpcSrv, healthService)

	// 注册HTTP服务
	api.RegisterHealthHTTPServer(httpSrv, healthService)

	return &Server{
		rpcSrv:        rpcSrv,
		httpSrv:       httpSrv,
		heartbeatSrv:  newHeartbeatServer(bc, healthService),
		consumerWatch: newConsumer(bc, data),
	}
}
