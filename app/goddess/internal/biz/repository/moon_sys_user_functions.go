package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysUserFunction 系统用户功能直授仓储接口
type SysUserFunction interface {
	// ListUserFunction 分页查询用户直授的功能列表（JOIN moon_sys_functions）
	ListUserFunction(ctx context.Context, req *bo.ListUserFunctionBo) ([]*bo.UserFunctionItemBo, int64, error)
	// GrantUserFunctions 批量直授功能（冲突时忽略，幂等）
	GrantUserFunctions(ctx context.Context, req *bo.GrantUserFunctionsBo) (int32, error)
	// RevokeUserFunctions 批量撤销直授功能（硬删除）
	RevokeUserFunctions(ctx context.Context, req *bo.RevokeUserFunctionsBo) (int32, error)
	// CountByUserUIDAndMenuCode 统计某用户在某菜单下拥有的直授功能数量
	CountByUserUIDAndMenuCode(ctx context.Context, userUID snowflake.ID, menuCode string) (int64, error)
}
