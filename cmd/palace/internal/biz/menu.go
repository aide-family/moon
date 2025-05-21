package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/job"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/server"
)

func NewMenuBiz(
	menuRepo repository.Menu,
	cacheRepo repository.Cache,
	logger log.Logger,
) *Menu {
	return &Menu{
		menuRepo:  menuRepo,
		cacheRepo: cacheRepo,
		helper:    log.NewHelper(log.With(logger, "module", "biz.menu")),
	}
}

type Menu struct {
	menuRepo  repository.Menu
	cacheRepo repository.Cache
	helper    *log.Helper
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
	return m.cacheRepo.GetMenu(ctx, operation)
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
	if err := m.menuRepo.ExistByName(ctx, menu.Name, menu.MenuId); err != nil {
		return err
	}
	if menu.MenuId == 0 {
		return m.menuRepo.Create(ctx, menu)
	}
	return m.menuRepo.Update(ctx, menu)
}

func (m *Menu) Jobs() []server.CronJob {
	return []server.CronJob{
		job.NewMenuJob(m.menuRepo, m.cacheRepo, m.helper.Logger()),
	}
}
