package biz

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
)

// NewServerRegisterBiz 创建服务器注册业务对象
func NewServerRegisterBiz(serverRegisterRepository microrepository.ServerRegister) *ServerRegisterBiz {
	return &ServerRegisterBiz{
		serverRegisterRepository: serverRegisterRepository,
	}
}

// ServerRegisterBiz 服务器注册业务对象
type ServerRegisterBiz struct {
	serverRegisterRepository microrepository.ServerRegister
}

// Heartbeat 心跳
func (s *ServerRegisterBiz) Heartbeat(ctx context.Context, request *api.HeartbeatRequest) error {
	return s.serverRegisterRepository.Heartbeat(ctx, request)
}

// GetServerList 获取服务器列表
func (s *ServerRegisterBiz) GetServerList(ctx context.Context, request *api.GetServerListRequest) (*api.GetServerListReply, error) {
	return s.serverRegisterRepository.GetServerList(ctx, request)
}
