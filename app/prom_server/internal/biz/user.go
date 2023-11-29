package biz

import (
	"context"
	"strconv"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/helper"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/helper/model/system"
	"prometheus-manager/pkg/util/password"
	"prometheus-manager/pkg/util/slices"
)

type (
	UserBiz struct {
		log *log.Helper

		userRepo  repository.UserRepo
		cacheRepo repository.CacheRepo
		roleRepo  repository.RoleRepo
	}
)

func NewUserBiz(userRepo repository.UserRepo, cacheRepo repository.CacheRepo, roleRepo repository.RoleRepo, logger log.Logger) *UserBiz {
	return &UserBiz{
		log: log.NewHelper(logger),

		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		roleRepo:  roleRepo,
	}
}

// GetUserInfoById 获取用户信息
func (b *UserBiz) GetUserInfoById(ctx context.Context, id uint32) (*dobo.UserBO, error) {
	user, err := b.userRepo.Get(ctx, system.UserInIds(id), system.UserPreloadRoles[uint32]())
	if err != nil {
		return nil, err
	}
	return dobo.NewUserDO(user).BO().First(), nil
}

// CreateUser 创建用户
func (b *UserBiz) CreateUser(ctx context.Context, user *dobo.UserBO) (*dobo.UserBO, error) {
	var err error
	userDo := dobo.NewUserBO(user).DO().First()
	userDo.Salt = password.GenerateSalt()
	userDo.Password, err = password.GeneratePassword(userDo.Password, userDo.Salt)
	if err != nil {
		return nil, err
	}

	userDo, err = b.userRepo.Create(ctx, userDo)
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

// LoginByUsernameAndPassword 登录
func (b *UserBiz) LoginByUsernameAndPassword(ctx context.Context, username, pwd string) (string, error) {
	userDo, err := b.userRepo.Get(ctx, system.UserEqName(username), system.UserPreloadRoles[uint32]())
	if err != nil {
		return "", err
	}

	if err = password.ValidatePasswordErr(pwd, userDo.Password, userDo.Salt); err != nil {
		return "", err
	}

	// 没有角色
	if len(userDo.Roles) == 0 {
		return middler.IssueToken(userDo.Id, "")
	}

	// 获取上次默认角色
	key := helper.UserRoleKey.KeyInt(userDo.Id).String()
	client, err := b.cacheRepo.Client()
	if err != nil {
		b.log.Error(err)
		return "", perrors.ErrorUnknown("系统错误")
	}

	cacheRoleIdStr := client.Get(ctx, key).String()
	searchRole := func(roleInfo *dobo.RoleDO) bool {
		cacheRoleId, _ := strconv.Atoi(cacheRoleIdStr)
		return roleInfo.Id == uint(cacheRoleId)
	}
	// 如果上次默认角色还在角色列表中
	if slices.ContainsOf(userDo.Roles, searchRole) {
		return middler.IssueToken(userDo.Id, cacheRoleIdStr)
	}

	roleId := userDo.Roles[0].Id
	roleIdStr := strconv.Itoa(int(roleId))

	return middler.IssueToken(userDo.Id, roleIdStr)
}

// Logout 退出登录
func (b *UserBiz) Logout(ctx context.Context, authClaims *middler.AuthClaims) error {
	client, err := b.cacheRepo.Client()
	if err != nil {
		return err
	}
	return middler.Expire(ctx, client, authClaims)
}

// RefreshToken 刷新token
func (b *UserBiz) RefreshToken(ctx context.Context, authClaims *middler.AuthClaims, roleId uint32) (string, error) {
	roleIdStr := strconv.Itoa(int(roleId))
	defer func() {
		key := helper.UserRoleKey.KeyInt(authClaims.ID).String()
		if err := b.cacheRepo.Set(ctx, key, roleIdStr, 0); err != nil {
			b.log.Errorf("cache user role err: %v", err)
			return
		}
	}()

	userDo, err := b.userRepo.Get(context.Background(), system.UserInIds(authClaims.ID), system.UserPreloadRoles[uint32]())
	if err != nil {
		return "", err
	}

	// 如果用户没有可用角色, 则直接置空处理
	if len(userDo.Roles) == 0 {
		roleIdStr = ""
		return middler.IssueToken(authClaims.ID, roleIdStr)
	}

	// 更改角色成功
	compareFun := func(do *dobo.RoleDO) bool {
		return do.Id == uint(roleId)
	}

	// 切换的角色不存在, 则检查已有角色和token内角色
	if !slices.ContainsOf(userDo.Roles, compareFun) {
		compareFunCurrRoleId := func(do *dobo.RoleDO) bool {
			currRoleId, _ := strconv.Atoi(authClaims.Role)
			return do.Id == uint(currRoleId)
		}
		// 先默认为token内的角色
		roleIdStr = authClaims.Role
		// 如果token的角色不在已有的角色列表中, 则默认第一个角色
		if !slices.ContainsOf(userDo.Roles, compareFunCurrRoleId) {
			roleIdStr = strconv.Itoa(int(userDo.Roles[0].Id))
		}
	}

	return middler.IssueToken(authClaims.ID, roleIdStr)
}

// EditUserPassword 修改密码
func (b *UserBiz) EditUserPassword(ctx context.Context, authClaims *middler.AuthClaims, oldPassword, newPassword string) (*dobo.UserBO, error) {
	userDo, err := b.userRepo.Get(ctx, system.UserInIds(authClaims.ID))
	if err != nil {
		return nil, err
	}
	// 验证旧密码
	if err = password.ValidatePasswordErr(oldPassword, userDo.Password, userDo.Salt); err != nil {
		return nil, err
	}

	// 加密新密码
	if userDo.Password, err = password.GeneratePassword(newPassword, userDo.Salt); err != nil {
		return nil, err
	}

	newUserDo := &dobo.UserDO{
		Id:       userDo.Id,
		Password: userDo.Password,
	}

	// 更新密码
	if _, err = b.userRepo.Update(ctx, newUserDo, system.UserInIds(userDo.Id)); err != nil {
		return nil, err
	}

	return dobo.NewUserDO(userDo).BO().First(), nil
}

// RelateRoles 关联角色
func (b *UserBiz) RelateRoles(ctx context.Context, userId uint32, roleIds []uint32) error {
	userDo, err := b.userRepo.Get(ctx, system.UserInIds(userId))
	if err != nil {
		return err
	}

	// 查询角色
	if len(roleIds) > 0 {
		roleDos, err := b.roleRepo.Find(ctx, system.RoleInIds(roleIds...))
		if err != nil {
			return err
		}
		userDo.Roles = roleDos
	}

	// 关联角色
	if err = b.userRepo.RelateRoles(ctx, userDo, userDo.Roles); err != nil {
		return err
	}

	return nil
}
