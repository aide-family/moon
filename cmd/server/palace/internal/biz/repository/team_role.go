package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/vobj"
)

type TeamRole interface {
	// CreateTeamRole 创建团队角色
	CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*bizmodel.SysTeamRole, error)

	// UpdateTeamRole 更新团队角色
	UpdateTeamRole(ctx context.Context, teamRole *bo.UpdateTeamRoleParams) error

	// DeleteTeamRole 删除团队角色
	DeleteTeamRole(ctx context.Context, id uint32) error

	// GetTeamRole 获取团队角色
	GetTeamRole(ctx context.Context, id uint32) (*bizmodel.SysTeamRole, error)

	// ListTeamRole 获取团队角色列表
	ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*bizmodel.SysTeamRole, error)

	// GetTeamRoleByUserID 获取用户团队角色
	GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*bizmodel.SysTeamRole, error)

	// UpdateTeamRoleStatus 更新团队角色状态
	UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error

	// CheckRbac 检查用户是否有权限
	CheckRbac(ctx context.Context, teamId uint32, roleIds []uint32, path string) (bool, error)
}
