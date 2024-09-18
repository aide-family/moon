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

// AuthorizationBiz 授权业务
type AuthorizationBiz struct {
	userRepo     repository.User
	teamRepo     repository.Team
	cacheRepo    repository.Cache
	teamRoleRepo repository.TeamRole
}

// NewAuthorizationBiz 创建授权业务
func NewAuthorizationBiz(
	userRepo repository.User,
	teamRepo repository.Team,
	cacheRepo repository.Cache,
	teamRoleRepo repository.TeamRole,
) *AuthorizationBiz {
	return &AuthorizationBiz{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		cacheRepo:    cacheRepo,
		teamRoleRepo: teamRoleRepo,
	}
}

// CheckPermission 检查用户是否有该资源权限
func (b *AuthorizationBiz) CheckPermission(ctx context.Context, req *bo.CheckPermissionParams) (*bizmodel.SysTeamMember, error) {
	if middleware.GetUserRole(ctx).IsAdminOrSuperAdmin() {
		return nil, nil
	}

	// 检查用户是否被团队禁用
	teamMemberDo, err := b.teamRepo.GetUserTeamByID(ctx, req.JwtClaims.GetUser(), req.JwtClaims.GetTeam())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nForbiddenUserNotInTeam(ctx)
		}
		return nil, err
	}
	if !teamMemberDo.Status.IsEnable() {
		return nil, merr.ErrorI18nForbiddenMemberDisabled(ctx)
	}

	if middleware.GetTeamRole(ctx).IsAdminOrSuperAdmin() {
		return teamMemberDo, nil
	}
	// 查询用户角色
	memberRoles, err := b.teamRoleRepo.GetTeamRoleByUserID(ctx, req.JwtClaims.GetUser(), req.JwtClaims.GetTeam())
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if len(memberRoles) == 0 {
		return nil, merr.ErrorI18nForbidden(ctx)
	}
	memberRoleIds := types.SliceTo(memberRoles, func(role *bizmodel.SysTeamRole) uint32 {
		return role.ID
	})
	rbac, err := b.teamRoleRepo.CheckRbac(ctx, req.JwtClaims.GetTeam(), memberRoleIds, req.Operation)
	if !types.IsNil(err) {
		return nil, err
	}
	if !rbac {
		return nil, merr.ErrorI18nForbidden(ctx).WithMetadata(map[string]string{
			"operation": req.Operation,
		})
	}

	return teamMemberDo, nil
}

// CheckToken 检查token
func (b *AuthorizationBiz) CheckToken(ctx context.Context, req *bo.CheckTokenParams) (*model.SysUser, error) {
	// 检查token是否过期
	if types.IsNil(req) || types.IsNil(req.JwtClaims) {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	if middleware.IsExpire(req.JwtClaims) {
		return nil, merr.ErrorI18nUnauthorizedJwtExpire(ctx)
	}
	// 检查token是否被登出
	if req.JwtClaims.IsLogout(ctx, b.cacheRepo.Cacher()) {
		return nil, merr.ErrorI18nUnauthorizedJwtBan(ctx)
	}

	// 检查用户是否被系统禁用
	userDo, err := b.userRepo.GetByID(ctx, req.JwtClaims.GetUser())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nUnauthorizedUserNotFound(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if !userDo.Status.IsEnable() {
		return nil, merr.ErrorI18nUnauthorizedUserBan(ctx)
	}
	return userDo, nil
}

// getJwtBaseInfo 获取jwtBaseInfo
func (b *AuthorizationBiz) getJwtBaseInfo(ctx context.Context, userDo *model.SysUser, teamID uint32) (*middleware.JwtBaseInfo, error) {
	if !userDo.Status.IsEnable() {
		return nil, merr.ErrorI18nUnauthorizedUserBan(ctx)
	}
	// 生成token
	base := new(middleware.JwtBaseInfo)
	base.SetUserInfo(userDo.ID)
	// 查询用户所属团队是否存在，存在着set temId memberId
	if teamID > 0 {
		memberItem, err := b.teamRepo.GetUserTeamByID(ctx, userDo.ID, teamID)
		if !types.IsNil(err) {
			return nil, merr.ErrorI18nForbiddenUserNotInTeam(ctx)
		}
		if !memberItem.Status.IsEnable() {
			return nil, merr.ErrorI18nForbiddenMemberDisabled(ctx)
		}
		base.SetTeamInfo(memberItem.TeamID)
		base.SetMember(memberItem.ID)
	}

	return base, nil
}

// Login 登录
func (b *AuthorizationBiz) Login(ctx context.Context, req *bo.LoginParams) (*bo.LoginReply, error) {
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByUsername(ctx, req.Username)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 统一包装成密码错误
			return nil, merr.ErrorI18nAlertPasswordErr(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	// 检查用户密码是否正确
	if err = checkPassword(ctx, userDo, req.Password); !types.IsNil(err) {
		return nil, err
	}

	// 生成token
	base, err := b.getJwtBaseInfo(ctx, userDo, req.Team)
	if !types.IsNil(err) {
		return nil, err
	}

	jwtClaims := middleware.NewJwtClaims(base)
	return &bo.LoginReply{
		JwtClaims: jwtClaims,
		User:      userDo,
	}, nil
}

// RefreshToken 刷新token
func (b *AuthorizationBiz) RefreshToken(ctx context.Context, req *bo.RefreshTokenParams) (*bo.RefreshTokenReply, error) {
	// 检查token是否过期
	if types.IsNil(req) || types.IsNil(req.JwtClaims) {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByID(ctx, req.JwtClaims.GetUser())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nUnauthorizedUserNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx)
	}

	// 生成token
	base, err := b.getJwtBaseInfo(ctx, userDo, req.Team)
	if !types.IsNil(err) {
		return nil, err
	}

	jwtClaims := middleware.NewJwtClaims(base)
	return &bo.RefreshTokenReply{
		User:      userDo,
		JwtClaims: jwtClaims,
	}, nil
}

// Logout 登出
func (b *AuthorizationBiz) Logout(ctx context.Context, params *bo.LogoutParams) error {
	return params.JwtClaims.Logout(ctx, b.cacheRepo.Cacher())
}

// 检查用户密码是否正确
func checkPassword(ctx context.Context, user *model.SysUser, password string) error {
	if err := types.ValidatePassword(user.Password, password, user.Salt); err != nil {
		return merr.ErrorI18nAlertPasswordErr(ctx).WithCause(err)
	}
	return nil
}
