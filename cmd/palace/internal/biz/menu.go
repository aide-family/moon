package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
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

func (m *Menu) GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error) {
	return m.menuRepo.GetMenuByOperation(ctx, operation)
}

func (m *Menu) GetMenu(ctx context.Context, id uint32) (do.Menu, error) {
	menus, err := m.menuRepo.Find(ctx, []uint32{id})
	if err != nil {
		return nil, err
	}
	if len(menus) == 0 {
		return nil, nil
	}
	return menus[0], nil
}

func (m *Menu) SaveMenu(ctx context.Context, menu *bo.SaveMenuRequest) error {
	if menu.MenuId == 0 {
		return m.menuRepo.Create(ctx, menu)
	}
	return m.menuRepo.Update(ctx, menu)
}
