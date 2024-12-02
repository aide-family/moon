package server

import (
	"github.com/aide-family/moon/api"
	alertapi "github.com/aide-family/moon/api/houyi/alert"
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

// Server 服务
type Server struct {
	rpcSrv             *grpc.Server
	httpSrv            *http.Server
	strategyWatch      *StrategyWatch
	alertWatch         *watch.Watcher
	heartbeatServer    *HeartbeatServer
	eventStrategyWatch *eventStrategyWatch
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
		s.alertWatch,
		s.heartbeatServer,
		s.eventStrategyWatch,
	}
}

// RegisterService 注册服务
func RegisterService(
	c *houyiconf.Bootstrap,
	data *data.Data,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	metricService *service.MetricService,
	healthService *service.HealthService,
	strategyService *service.StrategyService,
	alertService *service.AlertService,
) *Server {
	// 注册GRPC服务
	metadataapi.RegisterMetricServer(rpcSrv, metricService)
	api.RegisterHealthServer(rpcSrv, healthService)
	strategyapi.RegisterStrategyServer(rpcSrv, strategyService)
	alertapi.RegisterPushAlertServer(rpcSrv, alertService)
	api.RegisterAlertServer(rpcSrv, alertService)
	// 注册HTTP服务
	metadataapi.RegisterMetricHTTPServer(httpSrv, metricService)
	api.RegisterHealthHTTPServer(httpSrv, healthService)
	strategyapi.RegisterStrategyHTTPServer(httpSrv, strategyService)
	alertapi.RegisterPushAlertHTTPServer(httpSrv, alertService)
	api.RegisterAlertHTTPServer(httpSrv, alertService)

	return &Server{
		rpcSrv:             rpcSrv,
		httpSrv:            httpSrv,
		strategyWatch:      newStrategyWatch(c, data, alertService),
		alertWatch:         newAlertWatch(c, data, alertService),
		heartbeatServer:    newHeartbeatServer(c, healthService),
		eventStrategyWatch: newEventStrategyWatch(c, data, alertService),
	}
}
