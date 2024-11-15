package microserver

import (
	"context"
	"strings"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
)

// NewServerRegisterRepository 创建一个ServerRegisterRepository
func NewServerRegisterRepository(houyiConn *data.HouYiConn, rabbitConn *data.RabbitConn) microrepository.ServerRegister {
	return &serverRegisterRepositoryImpl{
		houyiConn:  houyiConn,
		rabbitConn: rabbitConn,
	}
}

type serverRegisterRepositoryImpl struct {
	houyiConn  *data.HouYiConn
	rabbitConn *data.RabbitConn
}

func (s *serverRegisterRepositoryImpl) GetServerList(ctx context.Context, request *api.GetServerListRequest) (*api.GetServerListReply, error) {
	if request.Type == "houyi" {
		return s.houyiConn.GetServerList()
	} else if request.Type == "rabbit" {
		return s.rabbitConn.GetServerList()
	} else {
		return &api.GetServerListReply{}, nil
	}

}

func (s *serverRegisterRepositoryImpl) Heartbeat(ctx context.Context, request *api.HeartbeatRequest) error {
	srv := request.GetServer()
	srvName := strings.ToLower(srv.GetName())
	switch srvName {
	case "houyi", "moon_houyi":
		return s.houyiConn.Heartbeat(ctx, request)
	case "rabbit", "moon_rabbit":
		return s.rabbitConn.Heartbeat(ctx, request)
	default:
		return merr.ErrorNotificationSystemError("服务不存在：%s", srv.GetName())
	}
}
