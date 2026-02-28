package biz

import (
	"context"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

func NewUser(userRepo repository.User, helper *klog.Helper) *User {
	return &User{
		userRepo: userRepo,
		helper:   klog.NewHelper(klog.With(helper.Logger(), "biz", "user")),
	}
}

type User struct {
	helper   *klog.Helper
	userRepo repository.User
}

func (u *User) GetUser(ctx context.Context, uid snowflake.ID) (*bo.UserItemBo, error) {
	user, err := u.userRepo.GetUser(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "get user failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get user %s failed", uid).WithCause(err)
	}
	return user, nil
}

func (u *User) ListUser(ctx context.Context, req *bo.ListUserBo) (*bo.PageResponseBo[*bo.UserItemBo], error) {
	page, err := u.userRepo.ListUser(ctx, req)
	if err != nil {
		u.helper.Errorw("msg", "list user failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list user failed").WithCause(err)
	}
	return page, nil
}

func (u *User) SelectUser(ctx context.Context, req *bo.SelectUserBo) (*bo.SelectUserBoResult, error) {
	result, err := u.userRepo.SelectUser(ctx, req)
	if err != nil {
		u.helper.Errorw("msg", "select user failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select user failed").WithCause(err)
	}
	return result, nil
}

func (u *User) BanUser(ctx context.Context, uid snowflake.ID) error {
	if err := u.userRepo.UpdateUserStatus(ctx, uid, enum.UserStatus_BANNED); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "ban user failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("ban user failed").WithCause(err)
	}
	return nil
}

func (u *User) PermitUser(ctx context.Context, uid snowflake.ID) error {
	if err := u.userRepo.UpdateUserStatus(ctx, uid, enum.UserStatus_ACTIVE); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "permit user failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("permit user failed").WithCause(err)
	}
	return nil
}

func (u *User) ChangeEmail(ctx context.Context, uid snowflake.ID, email string) error {
	if err := u.userRepo.UpdateUserEmail(ctx, uid, email); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "change email failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("change email failed").WithCause(err)
	}
	return nil
}

func (u *User) ChangeAvatar(ctx context.Context, uid snowflake.ID, avatar string) error {
	if err := u.userRepo.UpdateUserAvatar(ctx, uid, avatar); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user %s not found", uid)
		}
		u.helper.Errorw("msg", "change avatar failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("change avatar failed").WithCause(err)
	}
	return nil
}
