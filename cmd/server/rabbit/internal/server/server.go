package server

import (
	"context"

	"github.com/aide-family/moon/api"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	pushapi "github.com/aide-family/moon/api/rabbit/push"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/service"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, RegisterService)

// Server 服务
type Server struct {
	rpcSrv       *grpc.Server
	httpSrv      *http.Server
	heartbeatSrv *HeartbeatServer
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
	}
}

// RegisterService 注册服务
func RegisterService(
	bc *rabbitconf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	configService *service.ConfigService,
	hookService *service.HookService,
	healthService *service.HealthService,
) *Server {
	// 加载缓存配置
	if err := configService.LoadNotifyObject(context.Background()); !types.IsNil(err) {
		log.Errorw("加载配置失败", err)
	}
	// 注册GRPC服务
	pushapi.RegisterConfigServer(rpcSrv, configService)
	hookapi.RegisterHookServer(rpcSrv, hookService)
	api.RegisterHealthServer(rpcSrv, healthService)

	// 注册HTTP服务
	pushapi.RegisterConfigHTTPServer(httpSrv, configService)
	hookapi.RegisterHookHTTPServer(httpSrv, hookService)
	r := httpSrv.Route("/")
	r.POST("/v1/hook/send/{route}", hookService.HookSendMsgHTTPHandler())

	return &Server{
		rpcSrv:       rpcSrv,
		httpSrv:      httpSrv,
		heartbeatSrv: newHeartbeatServer(bc, healthService),
	}
}
