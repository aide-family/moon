package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Menu interface {
	Find(ctx context.Context, ids []uint32) ([]do.Menu, error)
	FindMenusByType(ctx context.Context, menuType vobj.MenuType) ([]do.Menu, error)
	GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error)
}
