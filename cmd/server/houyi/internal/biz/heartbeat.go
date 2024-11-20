package biz

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
)

// NewHeartbeatBiz 创建心跳业务
func NewHeartbeatBiz(heartbeatRepository repository.Heartbeat) *HeartbeatBiz {
	return &HeartbeatBiz{
		heartbeatRepository: heartbeatRepository,
	}
}

// HeartbeatBiz 心跳业务
type HeartbeatBiz struct {
	heartbeatRepository repository.Heartbeat
}

// Heartbeat 心跳
func (h *HeartbeatBiz) Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error {
	return h.heartbeatRepository.Heartbeat(ctx, in)
}
