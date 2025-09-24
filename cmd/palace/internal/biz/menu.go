package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/job"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/middler/permission"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

func NewMenuBiz(
	userRepo repository.User,
	memberRepo repository.Member,
	menuRepo repository.Menu,
	cacheRepo repository.Cache,
	logger log.Logger,
) *Menu {
	return &Menu{
		userRepo:   userRepo,
		memberRepo: memberRepo,
		menuRepo:   menuRepo,
		cacheRepo:  cacheRepo,
		helper:     log.NewHelper(log.With(logger, "module", "biz.menu")),
	}
}

type Menu struct {
	userRepo   repository.User
	memberRepo repository.Member
	menuRepo   repository.Menu
	cacheRepo  repository.Cache
	helper     *log.Helper
}

func (m *Menu) SelfMenus(ctx context.Context) ([]do.Menu, error) {
	userID := permission.GetUserIDByContextWithDefault(ctx, 0)
	if userID <= 0 {
		return nil, nil
	}
	userDo, err := m.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userDo.GetPosition().IsAdminOrSuperAdmin() {
		return m.menuRepo.FindAll(ctx)
	}
	menus := make([]do.Menu, 0, 100)
	for _, roleDo := range userDo.GetRoles() {
		menus = append(menus, roleDo.GetMenus()...)
	}
	userMenus, err := m.menuRepo.FindMenusByType(ctx, vobj.MenuTypeMenuUser)
	if err != nil {
		return nil, err
	}
	menus = append(menus, userMenus...)

	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok || teamID == 0 {
		return menus, nil
	}
	memberDo, err := m.memberRepo.FindByUserID(ctx, userID)
	if err != nil {
		if !merr.IsNotFound(err) {
			return nil, err
		}
		return menus, nil
	}
	if memberDo.GetPosition().IsAdminOrSuperAdmin() {
		teamMenus, err := m.menuRepo.FindMenusByType(ctx, vobj.MenuTypeMenuTeam)
		if err != nil {
			return nil, err
		}
		menus = append(menus, teamMenus...)
	} else {
		for _, roleDo := range memberDo.GetRoles() {
			menus = append(menus, roleDo.GetMenus()...)
		}
	}

	return menus, nil
}

func (m *Menu) Menus(ctx context.Context, params *bo.GetMenuTreeParams) ([]do.Menu, error) {
	return m.menuRepo.FindMenus(ctx, params)
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
	defer func() {
		apiPath := menu.APIPath
		if apiPath == "" {
			return
		}
		menu, err := m.menuRepo.GetMenuByOperation(ctx, apiPath)
		if err != nil {
			return
		}
		if err := m.cacheRepo.CacheMenus(ctx, menu); err != nil {
			m.helper.WithContext(ctx).Warnw("method", "menu.SaveMenu", "err", err)
		}
	}()
	if err := m.menuRepo.ExistByName(ctx, menu.Name, menu.MenuID); err != nil {
		return err
	}
	if menu.MenuID == 0 {
		return m.menuRepo.Create(ctx, menu)
	}
	return m.menuRepo.Update(ctx, menu)
}

func (m *Menu) Jobs() []cron_server.CronJob {
	return []cron_server.CronJob{
		job.NewMenuJob(m.menuRepo, m.cacheRepo, m.helper.Logger()),
	}
}

func (m *Menu) DeleteMenu(ctx context.Context, menuID uint32) error {
	return m.menuRepo.Delete(ctx, menuID)
}
