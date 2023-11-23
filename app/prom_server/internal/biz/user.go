package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/helper/model/system"
)

type (
	UserBiz struct {
		log *log.Helper

		userRepo repository.UserRepo
	}
)

func NewUserBiz(userRepo repository.UserRepo, logger log.Logger) *UserBiz {
	return &UserBiz{
		log: log.NewHelper(logger),

		userRepo: userRepo,
	}
}

// GetUserInfoById 获取用户信息
func (b *UserBiz) GetUserInfoById(ctx context.Context, id uint32) (*dobo.UserBO, error) {
	user, err := b.userRepo.Get(ctx, system.RoleInIds(id))
	if err != nil {
		return nil, err
	}
	return dobo.NewUserDO(user).BO().First(), nil
}

// CreateUser 创建用户
func (b *UserBiz) CreateUser(ctx context.Context, user *dobo.UserBO) (*dobo.UserBO, error) {
	userDo := dobo.NewUserBO(user).DO().First()
	userDo, err := b.userRepo.Create(ctx, userDo)
	if err != nil {
		return nil, err
	}
	return dobo.NewUserDO(userDo).BO().First(), nil
}

// UpdateUserById 更新用户信息
func (b *UserBiz) UpdateUserById(ctx context.Context, id uint32, user *dobo.UserBO) (*dobo.UserBO, error) {
	userDo := dobo.NewUserBO(user).DO().First()
	userDo, err := b.userRepo.Update(ctx, userDo, system.RoleInIds(id))
	if err != nil {
		return nil, err
	}
	return dobo.NewUserDO(userDo).BO().First(), nil
}

// UpdateUserStatusById 更新用户状态
func (b *UserBiz) UpdateUserStatusById(ctx context.Context, status valueobj.Status, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	userDo := &dobo.UserDO{Status: int32(status)}
	_, err := b.userRepo.Update(ctx, userDo, system.RoleInIds(ids...))
	return err
}

// DeleteUserByIds 删除用户
func (b *UserBiz) DeleteUserByIds(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return b.userRepo.Delete(ctx, system.RoleInIds(ids...))
}

// GetUserList 获取用户列表
func (b *UserBiz) GetUserList(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.UserBO, error) {
	userDos, err := b.userRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return dobo.NewUserDO(userDos...).BO().List(), nil
}
