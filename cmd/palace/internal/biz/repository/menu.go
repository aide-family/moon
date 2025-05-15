package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type Menu interface {
	Find(ctx context.Context, ids []uint32) ([]do.Menu, error)
}
