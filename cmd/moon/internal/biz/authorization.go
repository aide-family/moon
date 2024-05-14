package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/moon/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/types"
)

type AuthorizationBiz struct {
	userRepo  repo.UserRepo
	teamRepo  repo.TeamRepo
	cacheRepo repo.CacheRepo
}

func NewAuthorizationBiz(
	userRepo repo.UserRepo,
	teamRepo repo.TeamRepo,
	cacheRepo repo.CacheRepo,
) *AuthorizationBiz {
	return &AuthorizationBiz{
		userRepo:  userRepo,
		teamRepo:  teamRepo,
		cacheRepo: cacheRepo,
	}
}

// CheckPermission 检查用户是否有该资源权限
func (b *AuthorizationBiz) CheckPermission(ctx context.Context, req *bo.CheckPermissionParams) error {
	// 检查用户是否被团队禁用
	// 检查用户是否有该资源权限

	return nil
}

func (b *AuthorizationBiz) CheckToken(ctx context.Context, req *bo.CheckTokenParams) error {
	// 检查token是否过期
	if types.IsNil(req) || types.IsNil(req.JwtClaims) {
		return bo.UnLoginErr
	}
	if req.JwtClaims.VerifyExpiresAt(time.Now(), true) {
		return bo.UnLoginErr
	}
	// 检查token是否被登出
	// 检查用户是否被系统禁用
	return nil
}

// Login 登录
func (b *AuthorizationBiz) Login(ctx context.Context, req *bo.LoginParams) (*bo.LoginReply, error) {
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 统一包装成密码错误
			return nil, bo.PasswordErr
		}
		return nil, bo.SystemErr
	}
	// 检查用户密码是否正确
	if err = checkPassword(userDo, req.EnPassword); err != nil {
		return nil, err
	}

	// 生成token
	base := &middleware.JwtBaseInfo{}

	base.SetUserInfo(func() (userId, role uint32, err error) {
		return userDo.ID, uint32(userDo.Role), nil
	})
	base.SetTeamInfo(func() (teamId, teamRole uint32, err error) {
		if req.Team <= 0 {
			return
		}
		// 查询用户所属团队是否存在，存在着set temId
		_, err = b.teamRepo.GetUserTeamByID(ctx, userDo.ID, req.Team)
		if err != nil {
			return
		}
		teamId = req.Team
		if req.TeamRole <= 0 {
			return
		}
		// 查询用户所属团队角色是否存在，存在着set teamRoleId
		memberRoles, err := b.teamRepo.GetTeamRoleByUserID(ctx, userDo.ID, req.Team)
		if err != nil || len(memberRoles) == 0 {
			return
		}
		// 默认设置第一个角色
		teamRole = memberRoles[0].RoleID
		if req.TeamRole <= 0 {
			return
		}
		for _, role := range memberRoles {
			// 如果有设置角色，则设置该角色
			if role.RoleID == req.TeamRole {
				teamRole = role.RoleID
				break
			}
		}
		return
	})
	// 缓存token hash
	jwtClaims := middleware.NewJwtClaims(base)
	if err = jwtClaims.Cache(ctx, b.cacheRepo.Cacher()); err != nil {
		return nil, err
	}

	return &bo.LoginReply{
		JwtClaims: jwtClaims,
		User:      userDo,
	}, nil
}

// 检查用户密码是否正确
func checkPassword(user *model.SysUser, password string) error {
	_, err := types.DecryptPassword(password, user.Salt)
	if err == nil {
		return nil
	}
	return bo.PasswordErr
}
