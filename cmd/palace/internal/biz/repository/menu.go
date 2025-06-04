package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Menu interface {
	Find(ctx context.Context, ids []uint32) ([]do.Menu, error)
	FindAll(ctx context.Context, ids ...uint32) ([]do.Menu, error)
	FindMenusByType(ctx context.Context, menuType vobj.MenuType) ([]do.Menu, error)
	GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error)
	Create(ctx context.Context, menu *bo.SaveMenuRequest) error
	Update(ctx context.Context, menu *bo.SaveMenuRequest) error
	ExistByName(ctx context.Context, name string, menuID uint32) error
	List(ctx context.Context, req *bo.ListMenuParams) (*bo.ListMenuReply, error)
}
