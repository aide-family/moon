package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/vobj"
)

type Team interface {
	// GetUserTeamByID 查询用户指定团队信息
	GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*bizmodel.SysTeamMember, error)

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
	ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*bizmodel.SysTeamMember, error)

	// TransferTeamLeader 移交团队
	TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error
}
