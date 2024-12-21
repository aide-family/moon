package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// NewUserBiz 创建用户业务
func NewUserBiz(userRepo repository.User) *UserBiz {
	return &UserBiz{
		userRepo: userRepo,
	}
}

// UserBiz 用户业务
type UserBiz struct {
	userRepo repository.User
}

// CreateUser 创建用户
func (b *UserBiz) CreateUser(ctx context.Context, user *bo.CreateUserParams) (*model.SysUser, error) {
	if !middleware.GetUserRole(ctx).IsAdminOrSuperAdmin() {
		return nil, merr.ErrorI18nForbidden(ctx)
	}
	userDo, err := b.userRepo.Create(ctx, user)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return userDo, nil
}

// UpdateUser 更新用户
func (b *UserBiz) UpdateUser(ctx context.Context, user *bo.UpdateUserParams) error {
	if !middleware.GetUserRole(ctx).IsAdminOrSuperAdmin() {
		return merr.ErrorI18nForbidden(ctx)
	}
	// 记录操作日志
	if err := b.userRepo.UpdateByID(ctx, user); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateUserBaseInfo 更新用户基础信息
func (b *UserBiz) UpdateUserBaseInfo(ctx context.Context, user *bo.UpdateUserBaseParams) error {
	if err := b.userRepo.UpdateBaseByID(ctx, user); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// DeleteUser 删除用户
func (b *UserBiz) DeleteUser(ctx context.Context, id uint32) error {
	if !middleware.GetUserRole(ctx).IsAdminOrSuperAdmin() {
		return merr.ErrorI18nForbidden(ctx)
	}
	// 查询用户
	userDo, err := b.userRepo.GetByID(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastUserNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if !middleware.GetUserRole(ctx).IsSuperadmin() && userDo.Role.IsAdmin() {
		return merr.ErrorI18nForbidden(ctx)
	}
	// 记录操作日志
	log.Debugw("userDo", userDo)
	return b.userRepo.DeleteByID(ctx, id)
}

// GetUser 获取用户
func (b *UserBiz) GetUser(ctx context.Context, id uint32) (*model.SysUser, error) {
	userDo, err := b.userRepo.GetByID(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastUserNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return userDo, nil
}

// ListUser 获取用户列表
func (b *UserBiz) ListUser(ctx context.Context, params *bo.QueryUserListParams) ([]*model.SysUser, error) {
	userDos, err := b.userRepo.FindByPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return userDos, nil
}

// BatchUpdateUserStatus 批量更新用户状态
func (b *UserBiz) BatchUpdateUserStatus(ctx context.Context, params *bo.BatchUpdateUserStatusParams) error {
	if !middleware.GetUserRole(ctx).IsAdminOrSuperAdmin() {
		return merr.ErrorI18nForbidden(ctx)
	}
	// 不允许修改管理员状态
	// 查询所有用户详情
	userDos, err := b.userRepo.FindByIds(ctx, params.IDs...)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	for _, user := range userDos {
		if user.Role.IsAdmin() {
			return merr.ErrorI18nForbidden(ctx).WithMetadata(map[string]string{"msg": "不允许操作管理员状态"})
		}
	}

	if err := b.userRepo.UpdateStatusByIds(ctx, params.Status, params.IDs...); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// ResetUserPasswordBySelf 重置自己的密码
func (b *UserBiz) ResetUserPasswordBySelf(ctx context.Context, req *bo.ResetUserPasswordBySelfParams) error {
	if err := b.userRepo.UpdatePassword(ctx, req.UserID, req.Password); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// GetUserSelectList 获取用户下拉列表
func (b *UserBiz) GetUserSelectList(ctx context.Context, params *bo.QueryUserSelectParams) ([]*bo.SelectOptionBo, error) {
	userDos, err := b.userRepo.FindByPage(ctx, &bo.QueryUserListParams{
		Keyword: params.Keyword,
		Page:    params.Page,
		Status:  params.Status,
		Gender:  params.Gender,
		Role:    params.Role,
		IDs:     params.IDs,
	})
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return types.SliceTo(userDos, func(user *model.SysUser) *bo.SelectOptionBo {
		return bo.NewUserSelectOptionBuild(user).ToSelectOption()
	}), nil
}

// UpdateUserPhone 更新用户手机号
func (b *UserBiz) UpdateUserPhone(ctx context.Context, req *bo.UpdateUserPhoneRequest) error {
	userDo, err := b.userRepo.GetByID(ctx, req.UserID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastUserNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	userDo.Phone = req.Phone
	if err = b.userRepo.UpdateUser(ctx, userDo); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateUserEmail 更新用户邮箱
func (b *UserBiz) UpdateUserEmail(ctx context.Context, req *bo.UpdateUserEmailRequest) error {
	userDo, err := b.userRepo.GetByID(ctx, req.UserID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastUserNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	userDo.Email = req.Email
	if err = b.userRepo.UpdateUser(ctx, userDo); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateUserAvatar 更新用户头像
func (b *UserBiz) UpdateUserAvatar(ctx context.Context, req *bo.UpdateUserAvatarRequest) error {
	userDo, err := b.userRepo.GetByID(ctx, req.UserID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastUserNotFound(ctx)
		}
	}
	userDo.Avatar = req.Avatar
	if err = b.userRepo.UpdateUser(ctx, userDo); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// SetUserRole 设置用户角色
func (b *UserBiz) SetUserRole(ctx context.Context, params *bo.SetUserRoleParams) error {
	// 检查操作者角色
	opUserRole := middleware.GetUserRole(ctx)
	// 查询用户
	userDo, err := b.userRepo.GetByID(ctx, params.UserID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastUserNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if !opUserRole.GT(userDo.Role) {
		return merr.ErrorI18nForbiddenPermissionDenied(ctx).WithMetadata(map[string]string{"msg": "同等权限或者您的权限小于他", "me": opUserRole.String(), "other": userDo.Role.String()})
	}
	if opUserRole == params.Role {
		return merr.ErrorI18nForbiddenPermissionDenied(ctx).WithMetadata(map[string]string{"msg": "不能设置成同等权限"})
	}
	userDo.Role = params.Role
	userDo.WithContext(ctx)
	if err = b.userRepo.UpdateUser(ctx, userDo); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}
