package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewMenuBiz 菜单业务
func NewMenuBiz(teamMenuRepo repository.TeamMenu, msgRepo repository.Msg, menuRepo repository.Menu) *MenuBiz {
	return &MenuBiz{
		teamMenuRepo: teamMenuRepo,
		msgRepo:      msgRepo,
		menuRepo:     menuRepo,
	}
}

// MenuBiz 菜单业务
type MenuBiz struct {
	teamMenuRepo repository.TeamMenu
	menuRepo     repository.Menu
	msgRepo      repository.Msg
}

// MenuList 菜单列表
func (b *MenuBiz) MenuList(ctx context.Context) ([]*bizmodel.SysTeamMenu, error) {
	return b.teamMenuRepo.GetTeamMenuList(ctx, &bo.QueryTeamMenuListParams{TeamID: middleware.GetTeamID(ctx)})
}

// GetMenu 获取菜单详情
func (b *MenuBiz) GetMenu(ctx context.Context, menuID uint32) (*model.SysMenu, error) {
	menuDetail, err := b.menuRepo.GetByID(ctx, menuID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastMenuNotFound(ctx)
		}
		return nil, err
	}
	return menuDetail, nil
}

// BatchCreateMenu 批量创建菜单
func (b *MenuBiz) BatchCreateMenu(ctx context.Context, params []*bo.CreateMenuParams) error {
	err := b.menuRepo.BatchCreate(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateMenu 更新菜单
func (b *MenuBiz) UpdateMenu(ctx context.Context, params *bo.UpdateMenuParams) error {
	_, err := b.menuRepo.GetByID(ctx, params.ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastMenuNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return b.menuRepo.UpdateByID(ctx, params)
}

// UpdateMenuStatus 更新菜单状态
func (b *MenuBiz) UpdateMenuStatus(ctx context.Context, params *bo.UpdateMenuStatusParams) error {
	return b.menuRepo.UpdateStatusByIds(ctx, params.Status, params.IDs...)
}

// UpdateMenuTypes 更新菜单类型
func (b *MenuBiz) UpdateMenuTypes(ctx context.Context, params *bo.UpdateMenuTypeParams) error {
	return b.menuRepo.UpdateTypeByIds(ctx, params.Type, params.IDs...)
}

// ListMenuPage 分页菜单列表
func (b *MenuBiz) ListMenuPage(ctx context.Context, params *bo.QueryMenuListParams) ([]*model.SysMenu, error) {
	menuPage, err := b.menuRepo.FindByPage(ctx, params)
	if err != nil {
		return nil, err
	}
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return menuPage, nil
}

// DeleteMenu 删除菜单
func (b *MenuBiz) DeleteMenu(ctx context.Context, menuID uint32) error {
	_, err := b.menuRepo.GetByID(ctx, menuID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastMenuNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return b.menuRepo.DeleteByID(ctx, menuID)
}

// MenuAllList 获取所有菜单
func (b *MenuBiz) MenuAllList(ctx context.Context) ([]*model.SysMenu, error) {
	menus, err := b.menuRepo.ListAll(ctx)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return menus, nil
}
