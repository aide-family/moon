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

type AuthorizationBiz struct {
	userRepo     repository.User
	teamRepo     repository.Team
	cacheRepo    repository.Cache
	teamRoleRepo repository.TeamRole
}

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
func (b *AuthorizationBiz) CheckPermission(ctx context.Context, req *bo.CheckPermissionParams) error {
	if req.JwtClaims.GetTeamRole().IsAdmin() {
		return nil
	}
	// 检查用户是否被团队禁用
	teamMemberDo, err := b.teamRepo.GetUserTeamByID(ctx, req.JwtClaims.GetUser(), req.JwtClaims.GetTeam())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nUserNotInTeamErr(ctx)
		}
		return merr.ErrorI18nSystemErr(ctx)
	}
	if !teamMemberDo.Status.IsEnable() {
		return merr.ErrorI18nUserTeamDisabledErr(ctx)
	}

	// 查询用户角色
	memberRoles, err := b.teamRoleRepo.GetTeamRoleByUserID(ctx, req.JwtClaims.GetUser(), req.JwtClaims.GetTeam())
	if !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx)
	}
	if len(memberRoles) == 0 {
		return merr.ErrorI18nNoPermissionErr(ctx)
	}
	memberRoleIds := types.SliceTo(memberRoles, func(role *bizmodel.SysTeamRole) uint32 {
		return role.ID
	})
	rbac, err := b.teamRoleRepo.CheckRbac(ctx, req.JwtClaims.GetTeam(), memberRoleIds, req.Operation)
	if !types.IsNil(err) {
		return err
	}
	if !rbac {
		return merr.ErrorI18nNoPermissionErr(ctx).WithMetadata(map[string]string{
			"operation": req.Operation,
		})
	}

	return nil
}

func (b *AuthorizationBiz) CheckToken(ctx context.Context, req *bo.CheckTokenParams) error {
	// 检查token是否过期
	if types.IsNil(req) || types.IsNil(req.JwtClaims) {
		return merr.ErrorI18nUnLoginErr(ctx)
	}
	if middleware.IsExpire(req.JwtClaims) {
		return merr.ErrorI18nUnLoginErr(ctx)
	}
	// 检查token是否被登出
	if req.JwtClaims.IsLogout(ctx, b.cacheRepo.Cacher()) {
		return merr.ErrorI18nUnLoginErr(ctx)
	}

	// 检查用户是否被系统禁用
	userDo, err := b.userRepo.GetByID(ctx, req.JwtClaims.GetUser())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nUserNotFoundErr(ctx)
		}
		return merr.ErrorI18nSystemErr(ctx)
	}
	if !userDo.Status.IsEnable() {
		return merr.ErrorI18nUserLimitErr(ctx)
	}
	return nil
}

// getJwtBaseInfo 获取jwtBaseInfo
func (b *AuthorizationBiz) getJwtBaseInfo(ctx context.Context, userDo *model.SysUser, teamID uint32) (*middleware.JwtBaseInfo, error) {
	if !userDo.Status.IsEnable() {
		return nil, merr.ErrorI18nUserLimitErr(ctx)
	}
	// 生成token
	base := new(middleware.JwtBaseInfo)
	base.SetUserInfo(userDo)
	// 查询用户所属团队是否存在，存在着set temId
	if teamID > 0 {
		memberItem, err := b.teamRepo.GetUserTeamByID(ctx, userDo.ID, teamID)
		if !types.IsNil(err) {
			return nil, merr.ErrorI18nUserNotInTeamErr(ctx)
		}
		if !memberItem.Status.IsEnable() {
			return nil, merr.ErrorI18nUserTeamDisabledErr(ctx)
		}
		base.SetTeamInfo(memberItem)
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
			return nil, merr.ErrorI18nPasswordErr(ctx)
		}
		return nil, merr.ErrorI18nSystemErr(ctx)
	}
	// 检查用户密码是否正确
	if err = checkPassword(ctx, userDo, req.EnPassword); !types.IsNil(err) {
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
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByID(ctx, req.JwtClaims.GetUser())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nUserNotFoundErr(ctx)
		}
		return nil, merr.ErrorI18nSystemErr(ctx)
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
	decryptPassword, err := types.DecryptPassword(password, types.DefaultKey)
	if err != nil {
		return merr.ErrorI18nPasswordErr(ctx)
	}

	loginPass := types.NewPassword(decryptPassword, user.Salt)
	if loginPass.String() != user.Password {
		return merr.ErrorI18nPasswordErr(ctx)
	}
	return nil
}
