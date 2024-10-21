package repository

import (
	"context"

	"github.com/aide-family/moon/api"
)

// Heartbeat .
type Heartbeat interface {
	// Heartbeat 心跳包
	Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error
}
