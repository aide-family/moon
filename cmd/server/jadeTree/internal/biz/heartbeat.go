package biz

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/biz/repository"
)

// NewHeartbeatBiz 创建心跳包业务对象
func NewHeartbeatBiz(heartbeatRepository repository.Heartbeat) *HeartbeatBiz {
	return &HeartbeatBiz{
		heartbeatRepository: heartbeatRepository,
	}
}

// HeartbeatBiz 心跳包业务对象
type HeartbeatBiz struct {
	heartbeatRepository repository.Heartbeat
}

// Heartbeat 心跳包
func (b *HeartbeatBiz) Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error {
	return b.heartbeatRepository.Heartbeat(ctx, in)
}
