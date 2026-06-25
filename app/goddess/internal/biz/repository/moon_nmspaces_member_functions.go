package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesMemberFunction 租户成员功能直授仓储接口
type NmspacesMemberFunction interface {
	// ListNmspacesMemberFunction 分页查询成员直授的功能列表（JOIN moon_nmspaces_functions）
	ListNmspacesMemberFunction(ctx context.Context, req *bo.ListNmspacesMemberFunctionBo) ([]*bo.NmspacesMemberFunctionItemBo, int64, error)
	// GrantNmspacesMemberFunctions 批量直授功能（冲突时忽略，幂等）
	GrantNmspacesMemberFunctions(ctx context.Context, req *bo.GrantNmspacesMemberFunctionsBo) (int32, error)
	// RevokeNmspacesMemberFunctions 批量撤销直授功能（硬删除）
	RevokeNmspacesMemberFunctions(ctx context.Context, req *bo.RevokeNmspacesMemberFunctionsBo) (int32, error)
	// CountByNamespaceAndMemberUIDAndMenuCode 统计某成员在某菜单下拥有的直授功能数量
	CountByNamespaceAndMemberUIDAndMenuCode(ctx context.Context, namespaceUID, memberUID snowflake.ID, menuCode string) (int64, error)
}
