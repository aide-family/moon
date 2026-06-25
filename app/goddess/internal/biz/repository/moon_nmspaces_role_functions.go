package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesRoleFunction 租户角色功能权限仓储接口
type NmspacesRoleFunction interface {
	// ListNmspacesRoleFunction 分页查询角色的功能列表（JOIN moon_nmspaces_functions）
	ListNmspacesRoleFunction(ctx context.Context, req *bo.ListNmspacesRoleFunctionBo) ([]*bo.NmspacesRoleFunctionItemBo, int64, error)
	// GrantNmspacesRoleFunctions 批量授予功能（冲突时忽略，幂等）
	GrantNmspacesRoleFunctions(ctx context.Context, req *bo.GrantNmspacesRoleFunctionsBo) (int32, error)
	// RevokeNmspacesRoleFunctions 批量撤销功能（硬删除）
	RevokeNmspacesRoleFunctions(ctx context.Context, req *bo.RevokeNmspacesRoleFunctionsBo) (int32, error)
	// CountByNamespaceAndRoleCodeAndMenuCode 统计某角色在某菜单下拥有的功能数量
	CountByNamespaceAndRoleCodeAndMenuCode(ctx context.Context, namespaceUID snowflake.ID, roleCode, menuCode string) (int64, error)
	// ListFunctionsByNamespaceAndRoleCodes 查询多个角色拥有的所有功能（JOIN moon_nmspaces_functions，去重）
	ListFunctionsByNamespaceAndRoleCodes(ctx context.Context, namespaceUID snowflake.ID, roleCodes []string) ([]*bo.NmspacesRoleFunctionItemBo, error)
}
