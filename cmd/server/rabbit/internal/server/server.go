package server

import (
	"context"

	v1 "github.com/aide-family/moon/api/helloworld/v1"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	pushapi "github.com/aide-family/moon/api/rabbit/push"
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
	// 加载缓存配置
	if err := configService.LoadNotifyObject(context.Background()); !types.IsNil(err) {
		log.Errorw("加载配置失败", err)
	}
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
