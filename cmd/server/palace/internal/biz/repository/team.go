package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/vobj"
)

// Team 团队管理接口
type Team interface {
	// GetUserTeamByID 查询用户指定团队信息
	GetUserTeamByID(context.Context, uint32, uint32) (*bizmodel.SysTeamMember, error)

	// CreateTeam 创建团队
	CreateTeam(context.Context, *bo.CreateTeamParams) (*model.SysTeam, error)

	// UpdateTeam 更新团队信息
	UpdateTeam(context.Context, *bo.UpdateTeamParams) error

	// GetTeamDetail 获取团队详情
	GetTeamDetail(context.Context, uint32) (*model.SysTeam, error)

	// GetTeamList 获取团队列表
	GetTeamList(context.Context, *bo.QueryTeamListParams) ([]*model.SysTeam, error)

	// UpdateTeamStatus 修改团队状态
	UpdateTeamStatus(context.Context, vobj.Status, ...uint32) error

	// GetUserTeamList 获取用户团队列表
	GetUserTeamList(context.Context, uint32) ([]*model.SysTeam, error)

	// AddTeamMember 添加团队成员
	AddTeamMember(context.Context, *bo.AddTeamMemberParams) error

	// RemoveTeamMember 移除团队成员
	RemoveTeamMember(context.Context, *bo.RemoveTeamMemberParams) error

	// SetMemberAdmin 设置成员角色类型
	SetMemberAdmin(context.Context, *bo.SetMemberAdminParams) error

	// SetMemberRole 设置成员角色类型
	SetMemberRole(context.Context, *bo.SetMemberRoleParams) error

	// ListTeamMember 获取团队成员列表
	ListTeamMember(context.Context, *bo.ListTeamMemberParams) ([]*bizmodel.SysTeamMember, error)

	// TransferTeamLeader 移交团队
	TransferTeamLeader(context.Context, *bo.TransferTeamLeaderParams) error

	// GetTeamConfig 获取团队配置
	GetTeamConfig(context.Context, uint32) (*model.SysTeamConfig, error)

	// CreateTeamConfig 创建团队配置
	CreateTeamConfig(context.Context, *bo.SetTeamConfigParams) error

	// UpdateTeamConfig 更新团队配置
	UpdateTeamConfig(context.Context, *bo.SetTeamConfigParams) error

	// UpdateTeamMemberStatus 更新团队成员状态
	UpdateTeamMemberStatus(context.Context, vobj.Status, ...uint32) error

	// MemberList 成员列表
	MemberList(context.Context, uint32) ([]*bizmodel.SysTeamMember, error)

	// SyncTeamBaseData 同步团队基础数据
	SyncTeamBaseData(context.Context, *model.SysTeam, []*bizmodel.SysTeamMember) error

	// GetMemberDetail 获取团队成员详情
	GetMemberDetail(context.Context, uint32) (*bizmodel.SysTeamMember, error)

	// SyncTeamInfo 同步团队信息
	SyncTeamInfo(context.Context, ...uint32) error
}
