package biz

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
)

// NewHeartbeatBiz .
func NewHeartbeatBiz(heartbeatRepository repository.Heartbeat) *HeartbeatBiz {
	return &HeartbeatBiz{
		heartbeatRepository: heartbeatRepository,
	}
}

// HeartbeatBiz .
type HeartbeatBiz struct {
	heartbeatRepository repository.Heartbeat
}

func (h *HeartbeatBiz) Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error {
	return h.heartbeatRepository.Heartbeat(ctx, in)
}
