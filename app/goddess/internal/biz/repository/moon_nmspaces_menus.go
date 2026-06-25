package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesMenu 租户菜单仓储接口
type NmspacesMenu interface {
	// ListNmspacesMenu 分页查询租户菜单列表
	ListNmspacesMenu(ctx context.Context, req *bo.ListNmspacesMenuBo) ([]*bo.NmspacesMenuItemBo, int64, error)
	// GetNmspacesMenu 查询租户菜单详情
	GetNmspacesMenu(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (*bo.NmspacesMenuItemBo, error)
	// CreateNmspacesMenu 新增租户菜单
	CreateNmspacesMenu(ctx context.Context, req *bo.CreateNmspacesMenuBo) error
	// UpdateNmspacesMenu 编辑租户菜单
	UpdateNmspacesMenu(ctx context.Context, req *bo.UpdateNmspacesMenuBo) error
	// DeleteNmspacesMenu 删除租户菜单（软删除）
	DeleteNmspacesMenu(ctx context.Context, namespaceUID snowflake.ID, menuCode string) error
	// ExistsByNamespaceAndMenuCode 检查 menuCode 是否已存在（含软删除行）
	ExistsByNamespaceAndMenuCode(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (bool, error)
	// ExistsByNamespaceAndMenuCodeNotDeleted 检查 menuCode 是否在未删除行中存在
	ExistsByNamespaceAndMenuCodeNotDeleted(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (bool, error)
	// ExistsByNamespaceAndParentAndName 检查同一 parentMenuCode 下 menuName 是否已存在（未删除行）
	ExistsByNamespaceAndParentAndName(ctx context.Context, namespaceUID snowflake.ID, parentMenuCode, menuName string) (bool, error)
	// ExistsByNamespaceAndParentMenuCode 检查 parentMenuCode 是否存在且未删除
	ExistsByNamespaceAndParentMenuCode(ctx context.Context, namespaceUID snowflake.ID, parentMenuCode string) (bool, error)
	// CountChildren 统计子菜单数量（未删除行）
	CountChildren(ctx context.Context, namespaceUID snowflake.ID, parentMenuCode string) (int64, error)
}
