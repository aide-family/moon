package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/duke-git/lancet/v2/retry"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewInviteBiz 创建邀请业务对象
func NewInviteBiz(
	inviteRepo repository.TeamInvite,
	userRepo repository.User,
	teamRepo repository.Team,
	teamRoleRepo repository.TeamRole,
	userMessageRepo repository.UserMessage,
) *InviteBiz {
	return &InviteBiz{
		inviteRepo:      inviteRepo,
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		teamRoleRepo:    teamRoleRepo,
		userMessageRepo: userMessageRepo,
	}
}

type (
	// InviteBiz 邀请业务对象
	InviteBiz struct {
		inviteRepo      repository.TeamInvite
		userRepo        repository.User
		teamRepo        repository.Team
		teamRoleRepo    repository.TeamRole
		userMessageRepo repository.UserMessage
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

	params.UserID = user.ID
	params.TeamID = middleware.GetTeamID(ctx)
	// 校验邀请记录是否存在
	if err = i.checkInviteDataExists(ctx, params); !types.IsNil(err) {
		return err
	}

	// 获取邀请人信息
	opUser, err := i.userRepo.GetByID(ctx, middleware.GetUserID(ctx))
	if err != nil {
		return err
	}

	teamInvite, err := i.inviteRepo.InviteUser(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}

	return retry.Retry(func() error {
		return i.userMessageRepo.Create(ctx, &model.SysUserMessage{
			Content:  fmt.Sprintf("您收到一条来自 %s 的邀请，点击查看", opUser.Username),
			Category: vobj.UserMessageTypeInfo,
			UserID:   user.ID,
			Biz:      vobj.BizTypeInvitation,
			BizID:    teamInvite.ID,
		})
	}, retry.RetryTimes(3))
}

// UpdateInviteStatus 更新邀请状态
func (i *InviteBiz) UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error {
	teamInvite, err := i.inviteRepo.GetInviteDetail(ctx, params.InviteID)
	if !types.IsNil(err) {
		return merr.ErrorI18nToastTeamInviteNotFound(ctx)
	}
	opUserID := middleware.GetUserID(ctx)
	if opUserID != teamInvite.UserID {
		return merr.ErrorI18nToastTeamInviteNotFound(ctx)
	}

	if err := i.inviteRepo.UpdateInviteStatus(ctx, params); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	// 获取被邀请人
	inviter, err := i.userRepo.GetByID(ctx, teamInvite.UserID)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}

	return retry.Retry(func() error {
		return i.userMessageRepo.Create(ctx, &model.SysUserMessage{
			Content:  fmt.Sprintf("%s %s您的邀请，点击查看", inviter.Username, types.Ternary(params.InviteType.IsJoined(), "已同意", "已拒绝")),
			Category: types.Ternary(params.InviteType.IsJoined(), vobj.UserMessageTypeSuccess, vobj.UserMessageTypeError),
			UserID:   teamInvite.CreatorID,
			Biz:      types.Ternary(params.InviteType.IsJoined(), vobj.BizTypeInvitationAccepted, vobj.BizTypeInvitationRejected),
			BizID:    teamInvite.ID,
		})
	}, retry.RetryTimes(3), retry.RetryWithLinearBackoff(1*time.Second))
}

// UserInviteList 当前用户邀请列表
func (i *InviteBiz) UserInviteList(ctx context.Context, params *bo.QueryInviteListParams) ([]*model.SysTeamInvite, error) {
	inviteList, err := i.inviteRepo.UserInviteList(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return inviteList, nil
}

// TeamInviteDetail 团队邀请记录详情
func (i *InviteBiz) TeamInviteDetail(ctx context.Context, inviteID uint32) (*model.SysTeamInvite, error) {
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

// checkInviteDataExists 检查邀请数据是否存在
func (i *InviteBiz) checkInviteDataExists(ctx context.Context, params *bo.InviteUserParams) error {
	teamInvite, _ := i.inviteRepo.GetInviteUserByUserIDAndType(ctx, params)
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

	teamList, err := i.teamRepo.GetTeamList(ctx, &bo.QueryTeamListParams{IDs: teamIds})
	if err != nil {
		return teamMap
	}
	for _, team := range teamList {
		teamMap[team.ID] = team
	}
	return teamMap
}

// GetTeamRoles 获取团队角色
func (i *InviteBiz) GetTeamRoles(ctx context.Context, teamID uint32, roleIds []uint32) []*bizmodel.SysTeamRole {
	teamRoles, err := i.teamRoleRepo.GetBizTeamRolesByIds(ctx, teamID, roleIds)
	if !types.IsNil(err) {
		return nil
	}
	return teamRoles
}
