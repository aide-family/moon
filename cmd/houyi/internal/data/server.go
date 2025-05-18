package data

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"

	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/plugin/server"
)

type Server struct {
	network    config.Network
	rpcConn    *grpc.ClientConn
	httpClient *http.Client
}

func (s *Server) GetNetWork() config.Network {
	return s.network
}

func (s *Server) GetServerClient() common.ServerClient {
	return common.NewServerClient(s.rpcConn)
}

func (s *Server) GetServerHTTPClient() common.ServerHTTPClient {
	return common.NewServerHTTPClient(s.httpClient)
}

func (s *Server) GetCallbackClient() palace.CallbackClient {
	return palace.NewCallbackClient(s.rpcConn)
}

func (s *Server) GetCallbackHTTPClient() palace.CallbackHTTPClient {
	return palace.NewCallbackHTTPClient(s.httpClient)
}

func InitClient(initConfig *server.InitConfig) (*Server, error) {
	r := Server{
		network:    initConfig.MicroConfig.GetNetwork(),
		rpcConn:    nil,
		httpClient: nil,
	}
	switch r.network {
	case config.Network_GRPC:
		conn, err := server.InitGRPCClient(initConfig)
		if err != nil {
			return nil, err
		}
		r.rpcConn = conn
	case config.Network_HTTP:
		client, err := server.InitHTTPClient(initConfig)
		if err != nil {
			return nil, err
		}
		r.httpClient = client
	}
	return &r, nil
}
