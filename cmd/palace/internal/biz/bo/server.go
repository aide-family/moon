package bo

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/config"
)

type ServerRegisterReq struct {
	ServerType vobj.ServerType
	Server     *config.MicroServer
	Discovery  *config.Discovery
	TeamIds    []uint32
	IsOnline   bool
	Uuid       string
}

type Server struct {
	Config *ServerRegisterReq
	Conn   *grpc.ClientConn
	Client *http.Client
}

func (s *Server) Close() error {
	if s.IsGRPC() {
		return s.Conn.Close()
	}
	return s.Client.Close()
}

func (s *Server) IsGRPC() bool {
	return s.Config.Server.GetNetwork() == config.Network_GRPC
}

func (s *Server) IsHTTP() bool {
	return s.Config.Server.GetNetwork() == config.Network_HTTP
}
