package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewInviteBiz(inviteRepo repository.TeamInvite, userRepo repository.User) *InviteBiz {
	return &InviteBiz{
		inviteRepo: inviteRepo,
		userRepo:   userRepo,
	}
}

type (
	InviteBiz struct {
		inviteRepo repository.TeamInvite
		userRepo   repository.User
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
	if !types.IsNil(i.checkInviteDataExists(ctx, params)) {
		return err
	}

	err = i.inviteRepo.InviteUser(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}
func (i *InviteBiz) UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error {
	err := i.inviteRepo.UpdateInviteStatus(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

func (i *InviteBiz) InviteList(ctx context.Context, params *bo.QueryInviteListParams) ([]*bizmodel.SysTeamInvite, error) {
	inviteList, err := i.inviteRepo.InviteList(ctx)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return inviteList, nil
}

func (i *InviteBiz) checkInviteDataExists(ctx context.Context, params *bo.InviteUserParams) error {
	teamInvite, _ := i.inviteRepo.GetInviteUserByUserIdAndType(ctx, params)
	if !types.IsNil(teamInvite) {
		return merr.ErrorI18nToastTeamInviteAlreadyExists(ctx, params.InviteCode)
	}
	return nil
}
