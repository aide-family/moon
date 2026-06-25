package repository

import (
	"context"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysRoleMenu 系统角色菜单权限仓储接口
type SysRoleMenu interface {
	// ListRoleMenu 查询角色的菜单列表（JOIN moon_sys_menus）
	ListRoleMenu(ctx context.Context, req *bo.ListRoleMenuBo) ([]*bo.RoleMenuItemBo, error)
	// GrantRoleMenus 批量授予菜单（冲突时忽略，幂等）
	GrantRoleMenus(ctx context.Context, req *bo.GrantRoleMenusBo) (int32, error)
	// RevokeRoleMenus 批量撤销菜单（硬删除）
	RevokeRoleMenus(ctx context.Context, req *bo.RevokeRoleMenusBo) (int32, error)
	// ExistsByRoleCodeAndMenuCode 检查某菜单是否已授权给某角色
	ExistsByRoleCodeAndMenuCode(ctx context.Context, roleCode, menuCode string) (bool, error)
	// ListMenuCodesByRoleCodes 查询多个角色拥有的所有菜单编码（去重）
	ListMenuCodesByRoleCodes(ctx context.Context, roleCodes []string) ([]string, error)
}
