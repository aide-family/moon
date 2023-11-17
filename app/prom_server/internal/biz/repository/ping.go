package repository

import (
	"context"

	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

// PingRepo is a Greater repo.
type PingRepo interface {
	Ping(ctx context.Context, g *dobo.Ping) (*dobo.Ping, error)
}
