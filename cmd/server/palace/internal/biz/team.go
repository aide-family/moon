package biz

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewTeamBiz(teamRepo repository.Team) *TeamBiz {
	return &TeamBiz{
		teamRepo: teamRepo,
	}
}

type TeamBiz struct {
	teamRepo repository.Team
}

// CreateTeam 创建团队
func (t *TeamBiz) CreateTeam(ctx context.Context, params *bo.CreateTeamParams) (*model.SysTeam, error) {
	return t.teamRepo.CreateTeam(ctx, params)
}

// UpdateTeam 更新团队
func (t *TeamBiz) UpdateTeam(ctx context.Context, team *bo.UpdateTeamParams) error {
	// 不是管理员不允许修改
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsTeamAdminRole() {
		return bo.NoPermissionErr
	}
	return t.teamRepo.UpdateTeam(ctx, team)
}

// GetTeam 获取团队信息
func (t *TeamBiz) GetTeam(ctx context.Context, teamId uint32) (*model.SysTeam, error) {
	teamList, err := t.ListTeam(ctx, &bo.QueryTeamListParams{IDs: []uint32{teamId}})
	if !types.IsNil(err) {
		return nil, err
	}
	if len(teamList) == 0 {
		return nil, bo.TeamNotFoundErr
	}
	return teamList[0], nil
}

// ListTeam 获取团队列表
func (t *TeamBiz) ListTeam(ctx context.Context, params *bo.QueryTeamListParams) ([]*model.SysTeam, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, bo.UnLoginErr
	}
	if !claims.IsAdminRole() {
		params.UserID = claims.GetUser()
	}
	return t.teamRepo.GetTeamList(ctx, params)
}

// UpdateTeamStatus 更新团队状态
func (t *TeamBiz) UpdateTeamStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsAdminRole() && !claims.IsTeamAdminRole() {
		return bo.NoPermissionErr
	}
	return t.teamRepo.UpdateTeamStatus(ctx, status, ids...)
}

// GetUserTeamList 获取用户团队列表
func (t *TeamBiz) GetUserTeamList(ctx context.Context, userId uint32) ([]*model.SysTeam, error) {
	return t.teamRepo.GetUserTeamList(ctx, userId)
}

// AddTeamMember 添加团队成员
func (t *TeamBiz) AddTeamMember(ctx context.Context, params *bo.AddTeamMemberParams) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsTeamAdminRole() {
		return bo.NoPermissionErr
	}
	return t.teamRepo.AddTeamMember(ctx, params)
}

// RemoveTeamMember 移除团队成员
func (t *TeamBiz) RemoveTeamMember(ctx context.Context, params *bo.RemoveTeamMemberParams) error {
	if len(params.MemberIds) == 0 {
		return nil
	}
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsTeamAdminRole() {
		return bo.NoPermissionErr
	}
	// 查询团队管理员
	teamMemberList, err := t.teamRepo.ListTeamMember(ctx, &bo.ListTeamMemberParams{
		ID:        params.ID,
		MemberIDs: params.MemberIds,
	})
	if !types.IsNil(err) {
		return err
	}
	if len(teamMemberList) == 0 {
		return nil
	}
	// 查询团队信息
	teamInfo, err := t.teamRepo.GetTeamDetail(ctx, params.ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.TeamNotFoundErr
		}
		return err
	}

	for _, teamMember := range teamMemberList {
		role := vobj.Role(teamMember.Role)
		if role.IsSuperadmin() || role.IsAdmin() || teamMember.UserID == teamInfo.LeaderID {
			return bo.AdminUserDeleteErr
		}
		if teamMember.UserID == claims.GetUser() {
			return bo.DeleteSelfErr
		}
	}

	// 判断移除的人员中是否包含当前用户和管理员
	return t.teamRepo.RemoveTeamMember(ctx, params)
}

// SetTeamAdmin 设置团队管理员
func (t *TeamBiz) SetTeamAdmin(ctx context.Context, params *bo.SetMemberAdminParams) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.GetTeamRole().IsSuperadmin() {
		return bo.NoPermissionErr
	}
	// 不能设置自己
	for _, memberID := range params.MemberIds {
		if memberID == claims.GetUser() {
			return bo.TeamLeaderRepeatErr
		}
	}
	return t.teamRepo.SetMemberAdmin(ctx, params)
}

// SetMemberRole 设置团队成员角色
func (t *TeamBiz) SetMemberRole(ctx context.Context, params *bo.SetMemberRoleParams) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsTeamAdminRole() {
		return bo.NoPermissionErr
	}
	return t.teamRepo.SetMemberRole(ctx, params)
}

// ListTeamMember 获取团队成员列表
func (t *TeamBiz) ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*bizmodel.SysTeamMember, error) {
	return t.teamRepo.ListTeamMember(ctx, params)
}

// TransferTeamLeader 移交团队领导
func (t *TeamBiz) TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error {
	// 获取团队信息
	team, err := t.teamRepo.GetTeamDetail(ctx, params.ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.TeamNotFoundErr
		}
		return err
	}
	if team.LeaderID != params.OldLeaderID {
		return bo.TeamLeaderErr
	}
	if team.LeaderID == params.LeaderID {
		return bo.TeamLeaderRepeatErr
	}
	return t.teamRepo.TransferTeamLeader(ctx, params)
}
