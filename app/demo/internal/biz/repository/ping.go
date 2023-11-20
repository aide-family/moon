package repository

import (
	"context"

	"prometheus-manager/app/demo/internal/biz/dobo"
)

// PingRepo is a Greater repo.
type PingRepo interface {
	Ping(ctx context.Context, g *dobo.Ping) (*dobo.Ping, error)
}
