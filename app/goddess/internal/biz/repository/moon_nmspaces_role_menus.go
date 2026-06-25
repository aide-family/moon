package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesRoleMenu 租户角色菜单权限仓储接口
type NmspacesRoleMenu interface {
	// ListNmspacesRoleMenu 查询角色的菜单列表（JOIN moon_nmspaces_menus）
	ListNmspacesRoleMenu(ctx context.Context, req *bo.ListNmspacesRoleMenuBo) ([]*bo.NmspacesRoleMenuItemBo, error)
	// GrantNmspacesRoleMenus 批量授予菜单（冲突时忽略，幂等）
	GrantNmspacesRoleMenus(ctx context.Context, req *bo.GrantNmspacesRoleMenusBo) (int32, error)
	// RevokeNmspacesRoleMenus 批量撤销菜单（硬删除）
	RevokeNmspacesRoleMenus(ctx context.Context, req *bo.RevokeNmspacesRoleMenusBo) (int32, error)
	// ExistsByNamespaceAndRoleCodeAndMenuCode 检查某菜单是否已授权给某角色
	ExistsByNamespaceAndRoleCodeAndMenuCode(ctx context.Context, namespaceUID snowflake.ID, roleCode, menuCode string) (bool, error)
	// ListMenuCodesByNamespaceAndRoleCodes 查询多个角色拥有的所有菜单编码（去重）
	ListMenuCodesByNamespaceAndRoleCodes(ctx context.Context, namespaceUID snowflake.ID, roleCodes []string) ([]string, error)
}
