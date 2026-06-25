package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesMemberRole 租户成员角色权限仓储接口
type NmspacesMemberRole interface {
	// ListNmspacesMemberRole 分页查询成员的角色列表（JOIN moon_nmspaces_roles）
	ListNmspacesMemberRole(ctx context.Context, req *bo.ListNmspacesMemberRoleBo) ([]*bo.NmspacesMemberRoleItemBo, int64, error)
	// BindNmspacesMemberRoles 批量绑定成员角色（冲突时忽略，幂等）
	BindNmspacesMemberRoles(ctx context.Context, req *bo.BindNmspacesMemberRolesBo) (int32, error)
	// UnbindNmspacesMemberRoles 批量解绑成员角色（硬删除）
	UnbindNmspacesMemberRoles(ctx context.Context, req *bo.UnbindNmspacesMemberRolesBo) (int32, error)
	// CountByNamespaceAndMemberUID 统计某成员在当前租户拥有的角色数量
	CountByNamespaceAndMemberUID(ctx context.Context, namespaceUID, memberUID snowflake.ID) (int64, error)
	// CountByNamespaceAndMemberUIDAndRoleCodes 统计某成员在 roleCodes 中有几个角色
	CountByNamespaceAndMemberUIDAndRoleCodes(ctx context.Context, namespaceUID, memberUID snowflake.ID, roleCodes []string) (int64, error)
	// ListRoleCodesByNamespaceAndMemberUID 查询某成员在指定租户拥有的所有角色编码
	ListRoleCodesByNamespaceAndMemberUID(ctx context.Context, namespaceUID, memberUID snowflake.ID) ([]string, error)
}
