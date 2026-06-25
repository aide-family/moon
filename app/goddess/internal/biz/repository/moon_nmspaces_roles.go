package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesRole 租户角色仓储接口
type NmspacesRole interface {
	// ListNmspacesRole 分页查询租户角色列表
	ListNmspacesRole(ctx context.Context, req *bo.ListNmspacesRoleBo) (*bo.PageResponseBo[*bo.NmspacesRoleItemBo], error)
	// GetNmspacesRole 查询租户角色详情
	GetNmspacesRole(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (*bo.NmspacesRoleItemBo, error)
	// CreateNmspacesRole 新增租户角色
	CreateNmspacesRole(ctx context.Context, req *bo.CreateNmspacesRoleBo) error
	// UpdateNmspacesRole 编辑租户角色
	UpdateNmspacesRole(ctx context.Context, req *bo.UpdateNmspacesRoleBo) error
	// DeleteNmspacesRole 删除租户角色（软删除）
	DeleteNmspacesRole(ctx context.Context, namespaceUID snowflake.ID, roleCode string) error
	// ExistsByNamespaceAndRoleCode 检查 roleCode 是否已存在（含软删除行）
	ExistsByNamespaceAndRoleCode(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (bool, error)
	// ExistsByNamespaceAndRoleName 检查 roleName 是否在未删除行中已存在
	ExistsByNamespaceAndRoleName(ctx context.Context, namespaceUID snowflake.ID, roleName string) (bool, error)
	// ExistsByNamespaceAndRoleCodeNotDeleted 检查 roleCode 是否在未删除行中存在
	ExistsByNamespaceAndRoleCodeNotDeleted(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (bool, error)
}
