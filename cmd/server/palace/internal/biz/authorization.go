package biz

import (
	"context"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"

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
		//return nil
	}
	// 检查用户是否被团队禁用
	teamDo, err := b.teamRepo.GetUserTeamByID(ctx, req.JwtClaims.GetUser(), req.JwtClaims.GetTeam())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorModal("用户不在该团队中")
		}
		return merr.ErrorNotification("系统错误")
	}
	if !vobj.Status(teamDo.Status).IsEnable() {
		return merr.ErrorModal("用户被禁用")
	}

	// 查询用户角色
	memberRoles, err := b.teamRoleRepo.GetTeamRoleByUserID(ctx, req.JwtClaims.GetUser(), req.JwtClaims.GetTeam())
	if !types.IsNil(err) {
		return merr.ErrorNotification("系统错误")
	}
	if len(memberRoles) == 0 {
		return bo.NoPermissionErr
	}
	memberRoleIds := types.SliceTo(memberRoles, func(role *bizmodel.SysTeamRole) uint32 {
		return role.ID
	})
	rbac, err := b.teamRoleRepo.CheckRbac(ctx, req.JwtClaims.GetTeam(), memberRoleIds, req.Operation)
	if !types.IsNil(err) {
		return err
	}
	if !rbac {
		return bo.NoPermissionErr.WithMetadata(map[string]string{
			"operation": req.Operation,
			"rbac":      "false",
		})
	}

	return nil
}

func (b *AuthorizationBiz) CheckToken(ctx context.Context, req *bo.CheckTokenParams) error {
	// 检查token是否过期
	if types.IsNil(req) || types.IsNil(req.JwtClaims) {
		return bo.UnLoginErr
	}
	if middleware.IsExpire(req.JwtClaims) {
		return bo.UnLoginErr
	}
	// 检查token是否被登出
	if req.JwtClaims.IsLogout(ctx, b.cacheRepo.Cacher()) {
		return bo.UnLoginErr
	}

	// 检查用户是否被系统禁用
	userDo, err := b.userRepo.GetByID(ctx, req.JwtClaims.GetUser())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorModal("用户不存在")
		}
		return merr.ErrorNotification("系统错误")
	}
	if !userDo.Status.IsEnable() {
		return merr.ErrorModal("用户被禁用")
	}
	return nil
}

// Login 登录
func (b *AuthorizationBiz) Login(ctx context.Context, req *bo.LoginParams) (*bo.LoginReply, error) {
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByUsername(ctx, req.Username)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 统一包装成密码错误
			return nil, bo.PasswordErr
		}
		return nil, bo.SystemErr
	}
	// 检查用户密码是否正确
	if err = checkPassword(userDo, req.EnPassword); !types.IsNil(err) {
		return nil, err
	}

	// 生成token
	base := &middleware.JwtBaseInfo{}

	base.SetUserInfo(func() (userId uint32, role vobj.Role, err error) {
		return userDo.ID, userDo.Role, nil
	})
	base.SetTeamInfo(func() (teamId uint32, teamRole vobj.Role, err error) {
		if req.Team <= 0 {
			return
		}
		// 查询用户所属团队是否存在，存在着set temId
		memberItem, err := b.teamRepo.GetUserTeamByID(ctx, userDo.ID, req.Team)
		if !types.IsNil(err) {
			return 0, 0, err
		}
		return req.Team, vobj.Role(memberItem.Role), nil
	})

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
		return nil, bo.UnLoginErr
	}
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByID(ctx, req.JwtClaims.GetUser())
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 统一包装成密码错误
			return nil, merr.ErrorModal("用户不存在")
		}
		return nil, bo.SystemErr
	}
	if !userDo.Status.IsEnable() {
		return nil, merr.ErrorRedirect("用户被禁用").WithMetadata(map[string]string{
			"redirect": "/login",
		})
	}

	// 查询用户所属团队是否存在，存在着set temId
	teamMemberDo, err := b.teamRepo.GetUserTeamByID(ctx, userDo.ID, req.Team)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotification("用户不在该团队中")
		}
		return nil, bo.SystemErr
	}

	if !vobj.Status(teamMemberDo.Status).IsEnable() {
		return nil, merr.ErrorNotification("用户被禁用")
	}

	// 查询用户所属团队角色是否存在，存在着set teamRoleId
	memberItem, err := b.teamRepo.GetUserTeamByID(ctx, userDo.ID, req.Team)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotification("用户此权限已被移除")
		}
		return nil, bo.SystemErr
	}

	// 生成token
	base := &middleware.JwtBaseInfo{}
	base.SetUserInfo(func() (userId uint32, role vobj.Role, err error) {
		return userDo.ID, userDo.Role, nil
	})
	base.SetTeamInfo(func() (teamId uint32, teamRole vobj.Role, err error) {
		return req.Team, vobj.Role(memberItem.Role), nil
	})

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
func checkPassword(user *model.SysUser, password string) error {
	decryptPassword, err := types.DecryptPassword(password, types.DefaultKey)
	if err != nil {
		return bo.PasswordErr
	}

	loginPass := types.NewPassword(decryptPassword, user.Salt)
	if loginPass.String() != user.Password {
		return bo.PasswordErr
	}
	return nil
}
