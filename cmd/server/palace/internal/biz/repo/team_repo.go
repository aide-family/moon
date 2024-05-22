package repo

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type TeamRepo interface {
	// GetUserTeamByID 查询用户指定团队信息
	GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*model.SysTeamMember, error)

	// GetTeamRoleByUserID 查询用户指定团队角色
	GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*model.SysTeamMemberRole, error)

	// GetUserTeamRole 查询用户指定团队角色
	GetUserTeamRole(ctx context.Context, userID, teamID uint32) (*model.SysTeamMemberRole, error)

	// CreateTeam 创建团队
	CreateTeam(ctx context.Context, team *bo.CreateTeamParams) (*model.SysTeam, error)

	// UpdateTeam 更新团队信息
	UpdateTeam(ctx context.Context, team *bo.UpdateTeamParams) error

	// GetTeamDetail 获取团队详情
	GetTeamDetail(ctx context.Context, teamID uint32) (*model.SysTeam, error)

	// GetTeamList 获取团队列表
	GetTeamList(ctx context.Context, params *bo.QueryTeamListParams) ([]*model.SysTeam, error)

	// UpdateTeamStatus 修改团队状态
	UpdateTeamStatus(ctx context.Context, status vobj.Status, ids ...uint32) error

	// GetUserTeamList 获取用户团队列表
	GetUserTeamList(ctx context.Context, userID uint32) ([]*model.SysTeam, error)

	// AddTeamMember 添加团队成员
	AddTeamMember(ctx context.Context, params *bo.AddTeamMemberParams) error

	// RemoveTeamMember 移除团队成员
	RemoveTeamMember(ctx context.Context, params *bo.RemoveTeamMemberParams) error

	// SetMemberAdmin 设置成员角色类型
	SetMemberAdmin(ctx context.Context, params *bo.SetMemberAdminParams) error

	// SetMemberRole 设置成员角色类型
	SetMemberRole(ctx context.Context, params *bo.SetMemberRoleParams) error

	// ListTeamMember 获取团队成员列表
	ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*model.SysTeamMember, error)

	// TransferTeamLeader 移交团队
	TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error
}
