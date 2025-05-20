package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

func NewMenuBiz(menuRepo repository.Menu) *Menu {
	return &Menu{menuRepo: menuRepo}
}

type Menu struct {
	menuRepo repository.Menu
}

func (m *Menu) SelfMenus(ctx context.Context) ([]do.Menu, error) {
	return m.menuRepo.FindMenusByType(ctx, vobj.MenuTypeMenuUser)
}

func (m *Menu) TeamMenus(ctx context.Context) ([]do.Menu, error) {
	return m.menuRepo.FindMenusByType(ctx, vobj.MenuTypeMenuTeam)
}

func (m *Menu) SystemMenus(ctx context.Context) ([]do.Menu, error) {
	return m.menuRepo.FindMenusByType(ctx, vobj.MenuTypeMenuSystem)
}
