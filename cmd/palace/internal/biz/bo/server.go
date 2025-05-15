package bo

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/config"
)

type ServerRegisterReq struct {
	ServerType vobj.ServerType     `json:"server_type"`
	Server     *config.MicroServer `json:"server"`
	Discovery  *config.Discovery   `json:"discovery"`
	TeamIds    []uint32            `json:"team_ids"`
	IsOnline   bool                `json:"is_online"`
	Uuid       string              `json:"uuid"`
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
