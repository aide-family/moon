package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesMemberMenu 租户成员菜单直授仓储接口
type NmspacesMemberMenu interface {
	// ListNmspacesMemberMenu 查询成员直授的菜单列表（JOIN moon_nmspaces_menus）
	ListNmspacesMemberMenu(ctx context.Context, req *bo.ListNmspacesMemberMenuBo) ([]*bo.NmspacesMemberMenuItemBo, error)
	// GrantNmspacesMemberMenus 批量直授菜单（冲突时忽略，幂等）
	GrantNmspacesMemberMenus(ctx context.Context, req *bo.GrantNmspacesMemberMenusBo) (int32, error)
	// RevokeNmspacesMemberMenus 批量撤销直授菜单（硬删除）
	RevokeNmspacesMemberMenus(ctx context.Context, req *bo.RevokeNmspacesMemberMenusBo) (int32, error)
	// ExistsByNamespaceAndMemberUIDAndMenuCode 检查某菜单是否已直授给某成员
	ExistsByNamespaceAndMemberUIDAndMenuCode(ctx context.Context, namespaceUID, memberUID snowflake.ID, menuCode string) (bool, error)
}
