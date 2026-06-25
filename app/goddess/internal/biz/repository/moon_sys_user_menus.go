package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysUserMenu 系统用户菜单直授仓储接口
type SysUserMenu interface {
	// ListUserMenu 查询用户直授的菜单列表（JOIN moon_sys_menus）
	ListUserMenu(ctx context.Context, req *bo.ListUserMenuBo) ([]*bo.UserMenuItemBo, error)
	// GrantUserMenus 批量直授菜单（冲突时忽略，幂等）
	GrantUserMenus(ctx context.Context, req *bo.GrantUserMenusBo) (int32, error)
	// RevokeUserMenus 批量撤销直授菜单（硬删除）
	RevokeUserMenus(ctx context.Context, req *bo.RevokeUserMenusBo) (int32, error)
	// ExistsByUserUIDAndMenuCode 检查某菜单是否已直授给某用户
	ExistsByUserUIDAndMenuCode(ctx context.Context, userUID snowflake.ID, menuCode string) (bool, error)
}
