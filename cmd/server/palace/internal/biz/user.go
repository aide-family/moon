package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewUserBiz(userRepo repo.UserRepo, transactionRepo repo.TransactionRepo) *UserBiz {
	return &UserBiz{
		userRepo:        userRepo,
		TransactionRepo: transactionRepo,
	}
}

type UserBiz struct {
	userRepo repo.UserRepo
	repo.TransactionRepo
}

// CreateUser 创建用户
func (b *UserBiz) CreateUser(ctx context.Context, user *bo.CreateUserParams) (*model.SysUser, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, bo.UnLoginErr
	}
	if !claims.IsAdminRole() {
		return nil, bo.NoPermissionErr
	}
	userDo, err := b.userRepo.Create(ctx, user)
	if err != nil {
		return nil, bo.SystemErr.WithCause(err)
	}
	return userDo, nil
}

// UpdateUser 更新用户
func (b *UserBiz) UpdateUser(ctx context.Context, user *bo.UpdateUserParams) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsAdminRole() {
		return bo.NoPermissionErr
	}
	// 查询用户
	userDo, err := b.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.UserNotFoundErr
		}
		return bo.SystemErr.WithCause(err)
	}
	// 记录操作日志
	dePass, err := types.DecryptPassword(userDo.Password, userDo.Salt)
	if err != nil {
		return bo.SystemErr.WithCause(err)
	}
	user.Password = types.NewPassword(dePass, userDo.Salt)
	if err = b.userRepo.UpdateByID(ctx, user); err != nil {
		return bo.SystemErr.WithCause(err)
	}
	return nil
}

// DeleteUser 删除用户
func (b *UserBiz) DeleteUser(ctx context.Context, id uint32) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsAdminRole() {
		return bo.NoPermissionErr
	}
	// 查询用户
	userDo, err := b.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.UserNotFoundErr
		}
		return bo.SystemErr.WithCause(err)
	}
	if !claims.Role.IsSuperadmin() && vobj.Role(userDo.Role).IsAdmin() {
		return bo.AdminUserDeleteErr
	}
	// 记录操作日志
	log.Debugw("userDo", userDo)
	return b.userRepo.DeleteByID(ctx, id)
}

// GetUser 获取用户
func (b *UserBiz) GetUser(ctx context.Context, id uint32) (*model.SysUser, error) {
	userDo, err := b.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bo.UserNotFoundErr
		}
		return nil, bo.SystemErr.WithCause(err)
	}
	return userDo, nil
}

// ListUser 获取用户列表
func (b *UserBiz) ListUser(ctx context.Context, params *bo.QueryUserListParams) ([]*model.SysUser, error) {
	userDos, err := b.userRepo.FindByPage(ctx, params)
	if err != nil {
		return nil, bo.SystemErr.WithCause(err)
	}
	return userDos, nil
}

// BatchUpdateUserStatus 批量更新用户状态
func (b *UserBiz) BatchUpdateUserStatus(ctx context.Context, params *bo.BatchUpdateUserStatusParams) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	if !claims.IsAdminRole() {
		return bo.NoPermissionErr
	}
	// 不允许修改管理员状态
	// 查询所有用户详情
	if !claims.Role.IsSuperadmin() {
		userDos, err := b.userRepo.FindByIds(ctx, params.IDs...)
		if err != nil {
			return bo.SystemErr.WithCause(err)
		}
		for _, user := range userDos {
			if vobj.Role(user.Role).IsAdmin() {
				return bo.NoPermissionErr.WithMetadata(map[string]string{"msg": "不允许操作管理员状态"})
			}
		}
	}

	if err := b.userRepo.UpdateStatusByIds(ctx, params.Status, params.IDs...); err != nil {
		return bo.SystemErr.WithCause(err)
	}
	return nil
}

// ResetUserPasswordBySelf 重置自己的密码
func (b *UserBiz) ResetUserPasswordBySelf(ctx context.Context, req *bo.ResetUserPasswordBySelfParams) error {
	if err := b.userRepo.UpdatePassword(ctx, req.UserId, req.Password); err != nil {
		return bo.SystemErr.WithCause(err)
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
	})
	if err != nil {
		return nil, bo.SystemErr.WithCause(err)
	}
	return types.SliceTo(userDos, func(user *model.SysUser) *bo.SelectOptionBo {
		return bo.NewUserSelectOptionBuild(user).ToSelectOption()
	}), nil
}

// UpdateUserPhone 更新用户手机号
func (b *UserBiz) UpdateUserPhone(ctx context.Context, req *bo.UpdateUserPhoneRequest) error {
	userDo, err := b.userRepo.GetByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.UserNotFoundErr
		}
		return bo.SystemErr.WithCause(err)
	}
	userDo.Phone = req.Phone
	if err = b.userRepo.UpdateUser(ctx, userDo); err != nil {
		return bo.SystemErr.WithCause(err)
	}
	return nil
}

// UpdateUserEmail 更新用户邮箱
func (b *UserBiz) UpdateUserEmail(ctx context.Context, req *bo.UpdateUserEmailRequest) error {
	userDo, err := b.userRepo.GetByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.UserNotFoundErr
		}
		return bo.SystemErr.WithCause(err)
	}
	userDo.Email = req.Email
	if err = b.userRepo.UpdateUser(ctx, userDo); err != nil {
		return bo.SystemErr.WithCause(err)
	}
	return nil
}

// UpdateUserAvatar 更新用户头像
func (b *UserBiz) UpdateUserAvatar(ctx context.Context, req *bo.UpdateUserAvatarRequest) error {
	userDo, err := b.userRepo.GetByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.UserNotFoundErr
		}
	}
	userDo.Avatar = req.Avatar
	if err = b.userRepo.UpdateUser(ctx, userDo); err != nil {
		return bo.SystemErr.WithCause(err)
	}
	return nil
}
