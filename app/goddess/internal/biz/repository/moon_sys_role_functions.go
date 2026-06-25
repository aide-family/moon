package repository

import (
	"context"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysRoleFunction 系统角色功能权限仓储接口
type SysRoleFunction interface {
	// ListRoleFunction 分页查询角色的功能列表（JOIN moon_sys_functions）
	ListRoleFunction(ctx context.Context, req *bo.ListRoleFunctionBo) ([]*bo.RoleFunctionItemBo, int64, error)
	// GrantRoleFunctions 批量授予功能（冲突时忽略，幂等）
	GrantRoleFunctions(ctx context.Context, req *bo.GrantRoleFunctionsBo) (int32, error)
	// RevokeRoleFunctions 批量撤销功能（硬删除）
	RevokeRoleFunctions(ctx context.Context, req *bo.RevokeRoleFunctionsBo) (int32, error)
	// CountByRoleCodeAndMenuCode 统计某角色在某菜单下拥有的功能数量
	CountByRoleCodeAndMenuCode(ctx context.Context, roleCode, menuCode string) (int64, error)
	// ListFunctionsByRoleCodes 查询多个角色拥有的所有功能（JOIN moon_sys_functions，去重）
	ListFunctionsByRoleCodes(ctx context.Context, roleCodes []string) ([]*bo.RoleFunctionItemBo, error)
}
