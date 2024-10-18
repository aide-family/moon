package biz

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
)

// NewServerRegisterBiz .
func NewServerRegisterBiz(serverRegisterRepository microrepository.ServerRegister) *ServerRegisterBiz {
	return &ServerRegisterBiz{
		serverRegisterRepository: serverRegisterRepository,
	}
}

// ServerRegisterBiz .
type ServerRegisterBiz struct {
	serverRegisterRepository microrepository.ServerRegister
}

func (s *ServerRegisterBiz) Heartbeat(ctx context.Context, request *api.HeartbeatRequest) error {
	return s.serverRegisterRepository.Heartbeat(ctx, request)
}
