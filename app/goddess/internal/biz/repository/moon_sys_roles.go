package repository

import (
	"context"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysRole 系统角色仓储接口
type SysRole interface {
	// ListSysRole 分页查询系统角色列表
	ListSysRole(ctx context.Context, req *bo.ListSysRoleBo) (*bo.PageResponseBo[*bo.SysRoleItemBo], error)
	// GetSysRole 根据 roleCode 查询角色详情（含菜单和功能权限）
	GetSysRole(ctx context.Context, roleCode string) (*bo.GetSysRoleBo, error)
	// CreateSysRole 新增系统角色
	CreateSysRole(ctx context.Context, req *bo.CreateSysRoleBo) error
	// ExistsByRoleCode 检查 roleCode 是否已存在（含软删除行）
	ExistsByRoleCode(ctx context.Context, roleCode string) (bool, error)
	// ExistsByRoleName 检查 roleName 是否在未删除行中已存在
	ExistsByRoleName(ctx context.Context, roleName string) (bool, error)
	// UpdateSysRole 编辑系统角色
	UpdateSysRole(ctx context.Context, req *bo.UpdateSysRoleBo) error
	// DeleteSysRole 删除系统角色（软删除，role_status 置为 0）
	DeleteSysRole(ctx context.Context, roleCode string) error
	// ExistsByRoleCodeNotDeleted 检查 roleCode 是否在未删除行中存在
	ExistsByRoleCodeNotDeleted(ctx context.Context, roleCode string) (bool, error)
}
