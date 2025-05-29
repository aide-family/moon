package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/job"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/password"
	"github.com/aide-family/moon/pkg/util/validate"
)

// UserBiz is a user business logic implementation.
type UserBiz struct {
	userRepo  repository.User
	cacheRepo repository.Cache
	log       *log.Helper
}

// NewUserBiz creates a new UserBiz instance.
func NewUserBiz(
	userRepo repository.User,
	cacheRepo repository.Cache,
	logger log.Logger,
) *UserBiz {
	userBiz := &UserBiz{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		log:       log.NewHelper(log.With(logger, "module", "biz.user")),
	}
	do.RegisterGetUserFunc(func(id uint32) do.User {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return userBiz.getUser(ctx, id)
	})
	return userBiz
}

func (b *UserBiz) getUser(ctx context.Context, userID uint32) do.User {
	user, err := b.cacheRepo.GetUser(ctx, userID)
	if err != nil {
		if merr.IsUserNotFound(err) {
			user, err = b.userRepo.FindByID(ctx, userID)
			if err != nil {
				b.log.Errorw("msg", "get user fail", "err", err)
			} else {
				if err := b.cacheRepo.CacheUsers(ctx, user); err != nil {
					b.log.Errorw("msg", "cache user fail", "err", err)
				}
			}
		}
	}
	return user
}

// GetSelfInfo retrieves the current user's information from the context.
func (b *UserBiz) GetSelfInfo(ctx context.Context) (do.User, error) {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("user not found in context")
	}

	user, err := b.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, merr.ErrorInternalServer("failed to find user").WithCause(err)
	}

	return user, nil
}

// UpdateSelfInfo updates the current user's profile information.
func (b *UserBiz) UpdateSelfInfo(ctx context.Context, userUpdateInfo *bo.UserUpdateInfo) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}

	user, err := b.userRepo.FindByID(ctx, userID)
	if err != nil {
		return merr.ErrorInternalServer("failed to find user").WithCause(err)
	}

	if validate.IsNil(user) {
		return merr.ErrorUserNotFound("user not found")
	}

	if err = b.userRepo.UpdateUserInfo(ctx, userUpdateInfo.WithUser(user)); err != nil {
		return merr.ErrorInternalServer("failed to update user info").WithCause(err)
	}

	return nil
}

func (b *UserBiz) UpdateUserBaseInfo(ctx context.Context, userUpdateInfo *bo.UserUpdateInfo) error {
	user, err := b.userRepo.FindByID(ctx, userUpdateInfo.GetUserID())
	if err != nil {
		return err
	}
	if err = b.userRepo.UpdateUserInfo(ctx, userUpdateInfo.WithUser(user)); err != nil {
		return merr.ErrorInternalServer("failed to update user info").WithCause(err)
	}
	return nil
}

// UpdateSelfPassword updates the current user's password
func (b *UserBiz) UpdateSelfPassword(ctx context.Context, passwordUpdateInfo *bo.PasswordUpdateInfo) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}

	user, err := b.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if validate.IsNil(user) {
		return merr.ErrorUserNotFound("user not found")
	}

	// Verify old password
	if !user.ValidatePassword(passwordUpdateInfo.OldPassword) {
		return merr.ErrorPassword("old password is incorrect")
	}

	// Generate new password
	pass := password.New(passwordUpdateInfo.NewPassword)
	encryptedPassword, err := pass.EnValue()
	if err != nil {
		return merr.ErrorInternalServer("failed to encrypt password").WithCause(err)
	}

	updateUserPasswordInfo := &bo.UpdateUserPasswordInfo{
		UserID:         userID,
		Password:       encryptedPassword,
		Salt:           pass.Salt(),
		OriginPassword: passwordUpdateInfo.NewPassword,
		SendEmailFun:   passwordUpdateInfo.SendEmailFun,
	}
	// Update password through repository
	if err = b.userRepo.UpdatePassword(ctx, updateUserPasswordInfo); err != nil {
		return merr.ErrorInternalServer("failed to update password").WithCause(err)
	}

	return nil
}

// GetUserTeams retrieves all teams that the user is a member of
func (b *UserBiz) GetUserTeams(ctx context.Context) ([]do.Team, error) {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("user not found in context")
	}

	teams, err := b.userRepo.GetTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, merr.ErrorInternalServer("failed to get user teams").WithCause(err)
	}

	return teams, nil
}

func (b *UserBiz) UpdateUserStatus(ctx context.Context, req *bo.UpdateUserStatusRequest) error {
	return b.userRepo.UpdateUserStatus(ctx, req)
}

func (b *UserBiz) ResetUserPassword(ctx context.Context, req *bo.ResetUserPasswordRequest) error {
	user, err := b.userRepo.FindByID(ctx, req.UserId)
	if err != nil {
		return err
	}
	newPass := password.GenerateRandomPassword(8)
	pass := password.New(newPass)
	enValue, err := pass.EnValue()
	if err != nil {
		return err
	}
	updateUserPasswordInfo := &bo.UpdateUserPasswordInfo{
		UserID:         user.GetID(),
		Password:       enValue,
		Salt:           pass.Salt(),
		OriginPassword: newPass,
		SendEmailFun:   req.SendEmailFun,
	}
	return b.userRepo.UpdatePassword(ctx, updateUserPasswordInfo)
}

func (b *UserBiz) UpdateUserPosition(ctx context.Context, req *bo.UpdateUserPositionRequest) error {
	operatorID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operator, err := b.userRepo.FindByID(ctx, operatorID)
	if err != nil {
		return err
	}
	req.WithOperator(operator)
	user, err := b.userRepo.FindByID(ctx, req.UserId)
	if err != nil {
		return err
	}
	req.WithUser(user)
	if err := req.Validate(); err != nil {
		return err
	}
	return b.userRepo.UpdateUserPosition(ctx, req)
}

func (b *UserBiz) GetUser(ctx context.Context, userID uint32) (do.User, error) {
	return b.userRepo.FindByID(ctx, userID)
}

func (b *UserBiz) ListUser(ctx context.Context, req *bo.UserListRequest) (*bo.UserListReply, error) {
	return b.userRepo.List(ctx, req)
}

func (b *UserBiz) Jobs() []cron_server.CronJob {
	return []cron_server.CronJob{
		job.NewUserJob(b.userRepo, b.cacheRepo, b.log.Logger()),
	}
}
