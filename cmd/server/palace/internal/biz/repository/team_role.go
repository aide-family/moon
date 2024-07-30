package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/vobj"
)

// TeamRole 团队角色接口
type TeamRole interface {
	// CreateTeamRole 创建团队角色
	CreateTeamRole(context.Context, *bo.CreateTeamRoleParams) (*bizmodel.SysTeamRole, error)

	// UpdateTeamRole 更新团队角色
	UpdateTeamRole(context.Context, *bo.UpdateTeamRoleParams) error

	// DeleteTeamRole 删除团队角色
	DeleteTeamRole(context.Context, uint32) error

	// GetTeamRole 获取团队角色
	GetTeamRole(context.Context, uint32) (*bizmodel.SysTeamRole, error)

	// ListTeamRole 获取团队角色列表
	ListTeamRole(context.Context, *bo.ListTeamRoleParams) ([]*bizmodel.SysTeamRole, error)

	// GetTeamRoleByUserID 获取用户团队角色
	GetTeamRoleByUserID(context.Context, uint32, uint32) ([]*bizmodel.SysTeamRole, error)

	// UpdateTeamRoleStatus 更新团队角色状态
	UpdateTeamRoleStatus(context.Context, vobj.Status, ...uint32) error

	// CheckRbac 检查用户是否有权限
	CheckRbac(context.Context, uint32, []uint32, string) (bool, error)
}
