package repository

import (
	"context"

	"github.com/aide-family/moon/api"
)

// Heartbeat .
type Heartbeat interface {
	Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error
}
