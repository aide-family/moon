package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewInviteBiz 创建InviteBiz
func NewInviteBiz(inviteRepo repository.TeamInvite, userRepo repository.User, teamRepo repository.Team) *InviteBiz {
	return &InviteBiz{
		inviteRepo: inviteRepo,
		userRepo:   userRepo,
		teamRepo:   teamRepo,
	}
}

type (
	InviteBiz struct {
		inviteRepo repository.TeamInvite
		userRepo   repository.User
		teamRepo   repository.Team
	}
)

// InviteUser 邀请用户
func (i *InviteBiz) InviteUser(ctx context.Context, params *bo.InviteUserParams) error {
	user, err := i.userRepo.GetUserByEmailOrPhone(ctx, params.InviteCode)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastUserNotFound(ctx)
		}
		return err
	}
	// 获取团队id
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return merr.ErrorI18nUnauthorized(ctx)
	}

	params.UserID = user.ID
	params.TeamID = claims.TeamID
	// 校验邀请记录是否存在
	if err = i.checkInviteDataExists(ctx, params); !types.IsNil(err) {
		return err
	}

	err = i.inviteRepo.InviteUser(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateInviteStatus 更新邀请状态
func (i *InviteBiz) UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error {
	err := i.inviteRepo.UpdateInviteStatus(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UserInviteList 当前用户邀请列表
func (i *InviteBiz) UserInviteList(ctx context.Context, params *bo.QueryInviteListParams) ([]*bizmodel.SysTeamInvite, error) {
	inviteList, err := i.inviteRepo.UserInviteList(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return inviteList, nil
}

// TeamInviteDetail 团队邀请记录详情
func (i *InviteBiz) TeamInviteDetail(ctx context.Context, inviteID uint32) (*bizmodel.SysTeamInvite, error) {
	detail, err := i.inviteRepo.GetInviteDetail(ctx, inviteID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastTeamInviteNotFound(ctx)
		}
		return nil, err
	}
	return detail, nil
}

// DeleteInvite 删除邀请记录
func (i *InviteBiz) DeleteInvite(ctx context.Context, inviteID uint32) error {
	err := i.inviteRepo.DeleteInvite(ctx, inviteID)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

func (i *InviteBiz) checkInviteDataExists(ctx context.Context, params *bo.InviteUserParams) error {
	teamInvite, _ := i.inviteRepo.GetInviteUserByUserIdAndType(ctx, params)
	if !types.IsNil(teamInvite) && !teamInvite.InviteType.IsRejected() {
		return merr.ErrorI18nToastTeamInviteAlreadyExists(ctx, params.InviteCode)
	}
	// 检查用户是否已经加入团队
	teamMember, _ := i.teamRepo.GetUserTeamByID(ctx, params.UserID, params.TeamID)
	if !types.IsNil(teamMember) {
		return merr.ErrorI18nToastTeamInviteAlreadyExists(ctx, params.InviteCode)
	}
	return nil
}

// GetTeamMapByIds 根据ID获取团队信息
func (i *InviteBiz) GetTeamMapByIds(ctx context.Context, teamIds []uint32) map[uint32]*model.SysTeam {
	teamMap := make(map[uint32]*model.SysTeam)

	param := &bo.QueryTeamListParams{
		IDs: teamIds,
	}

	teamList, err := i.teamRepo.GetTeamList(ctx, param)
	if err != nil {
		return teamMap
	}
	for _, team := range teamList {
		teamMap[team.ID] = team
	}
	return teamMap
}
