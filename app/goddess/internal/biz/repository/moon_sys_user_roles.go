package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysUserRole 系统用户角色权限仓储接口
type SysUserRole interface {
	// ListSysUserRole 分页查询用户的系统角色列表（JOIN moon_sys_roles）
	ListSysUserRole(ctx context.Context, req *bo.ListSysUserRoleBo) ([]*bo.SysUserRoleItemBo, int64, error)
	// BindSysUserRoles 批量绑定用户角色（冲突时忽略，幂等）
	BindSysUserRoles(ctx context.Context, req *bo.BindSysUserRolesBo) (int32, error)
	// UnbindSysUserRoles 批量解绑用户角色（硬删除）
	UnbindSysUserRoles(ctx context.Context, req *bo.UnbindSysUserRolesBo) (int32, error)
	// CountByUserUID 统计某用户当前拥有的角色数量
	CountByUserUID(ctx context.Context, userUID snowflake.ID) (int64, error)
	// CountByUserUIDAndRoleCodes 统计某用户在 roleCodes 中有几个角色
	CountByUserUIDAndRoleCodes(ctx context.Context, userUID snowflake.ID, roleCodes []string) (int64, error)
	// ListRoleCodesByUserUID 查询某用户拥有的所有角色编码
	ListRoleCodesByUserUID(ctx context.Context, userUID snowflake.ID) ([]string, error)
}
