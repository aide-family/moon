package microrepository

import (
	"context"

	"github.com/aide-family/moon/api"
)

// ServerRegister 子服务注册方法
type ServerRegister interface {
	// Heartbeat 心跳
	Heartbeat(context.Context, *api.HeartbeatRequest) error
}
