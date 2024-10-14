package biz

import (
	"context"
	"fmt"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo/auth"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/helper"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// AuthorizationBiz 授权业务
type AuthorizationBiz struct {
	userRepo     repository.User
	teamRepo     repository.Team
	cacheRepo    repository.Cache
	teamRoleRepo repository.TeamRole
	resourceRepo repository.Resource
	oAuthRepo    repository.OAuth

	githubOAuthConf *oauth2.Config
	giteeOAuthConf  *oauth2.Config
	redirectURL     string
}

// NewAuthorizationBiz 创建授权业务
func NewAuthorizationBiz(
	bc *palaceconf.Bootstrap,
	userRepo repository.User,
	teamRepo repository.Team,
	cacheRepo repository.Cache,
	teamRoleRepo repository.TeamRole,
	resourceRepo repository.Resource,
	oAuthRepo repository.OAuth,
) *AuthorizationBiz {
	githubOAuthConf := bc.GetOauth2().GetGithub()
	giteeOAuthConf := bc.GetOauth2().GetGitee()
	return &AuthorizationBiz{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		cacheRepo:    cacheRepo,
		teamRoleRepo: teamRoleRepo,
		resourceRepo: resourceRepo,
		oAuthRepo:    oAuthRepo,
		githubOAuthConf: &oauth2.Config{
			ClientID:     githubOAuthConf.GetClientId(),
			ClientSecret: githubOAuthConf.GetClientSecret(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
			RedirectURL: githubOAuthConf.GetCallbackUri(),
			Scopes:      githubOAuthConf.GetScopes(),
		},
		giteeOAuthConf: &oauth2.Config{
			ClientID:     giteeOAuthConf.GetClientId(),
			ClientSecret: giteeOAuthConf.GetClientSecret(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://gitee.com/oauth/authorize",
				TokenURL: "https://gitee.com/oauth/token",
			},
			RedirectURL: giteeOAuthConf.GetCallbackUri(),
			Scopes:      giteeOAuthConf.GetScopes(),
		},
		redirectURL: bc.GetOauth2().GetRedirectUri(),
	}
}

// CheckPermission 检查用户是否有该资源权限
func (b *AuthorizationBiz) CheckPermission(ctx context.Context, req *bo.CheckPermissionParams) (*bizmodel.SysTeamMember, error) {
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

	if teamMemberDo.Role.IsAdminOrSuperAdmin() {
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
	memberRoleIds := types.SliceToWithFilter(memberRoles, func(role *bizmodel.SysTeamRole) (uint32, bool) {
		return role.ID, role.Status.IsEnable()
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
	// 查询接口是否停用
	if err := b.resourceRepo.CheckPath(ctx, req.Operation); err != nil {
		return nil, err
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
		base.SetTeamInfo(teamID)
		base.SetMember(memberItem.ID)
	}

	return base, nil
}

// Login 登录
func (b *AuthorizationBiz) Login(ctx context.Context, req *bo.LoginParams) (*bo.LoginReply, error) {
	// 检查用户是否存在
	userDo, err := b.userRepo.GetByEmail(ctx, req.Username)
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
	if err := types.NewPassword(user.Password, user.Salt).Validate(password); err != nil {
		return merr.ErrorI18nAlertPasswordErr(ctx).WithCause(err)
	}
	return nil
}

// GetOAuthConf 获取oauth配置
func (b *AuthorizationBiz) GetOAuthConf(provider vobj.OAuthAPP) *oauth2.Config {
	switch provider {
	case vobj.OAuthAPPGithub:
		return b.githubOAuthConf
	case vobj.OAuthAPPGitee:
		return b.giteeOAuthConf
	default:
		return nil
	}
}

// OAuthLogin oauth登录
func (b *AuthorizationBiz) OAuthLogin(ctx context.Context, provider vobj.OAuthAPP, code string) (string, error) {
	switch provider {
	case vobj.OAuthAPPGithub:
		return b.githubLogin(ctx, code)
	case vobj.OAuthAPPGitee:
		return b.giteeLogin(ctx, code)
	default:
		return "", merr.ErrorI18nNotificationSystemError(ctx)
	}
}

// githubLogin github登录
func (b *AuthorizationBiz) githubLogin(ctx context.Context, code string) (string, error) {
	token, err := b.githubOAuthConf.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	// 使用token来获取用户信息
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return "", err
	}
	body := userResp.Body
	defer body.Close()
	var userInfo auth.GithubUser
	if err := types.NewDecoder(body).Decode(&userInfo); err != nil {
		return "", err
	}

	//fmt.Println(userInfo.String())
	return b.oauthLogin(ctx, &userInfo)
}

// giteeLogin gitee登录
func (b *AuthorizationBiz) giteeLogin(ctx context.Context, code string) (string, error) {
	opts := []oauth2.AuthCodeOption{
		// https://gitee.com/oauth/token?grant_type=authorization_code&code={code}&client_id={client_id}&redirect_uri={redirect_uri}&client_secret={client_secret}
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("client_secret", b.giteeOAuthConf.ClientSecret),
		oauth2.SetAuthURLParam("client_id", b.giteeOAuthConf.ClientID),
		oauth2.SetAuthURLParam("redirect_uri", b.giteeOAuthConf.RedirectURL),
		oauth2.SetAuthURLParam("code", code),
	}
	token, err := b.giteeOAuthConf.Exchange(context.Background(), code, opts...)
	if err != nil {
		return "", err
	}
	// 使用token来获取用户信息
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

	resp, err := client.Get("https://gitee.com/api/v5/user")
	if err != nil {
		return "", err
	}
	body := resp.Body
	defer body.Close()
	var userInfo auth.GiteeUser
	if err := types.NewDecoder(body).Decode(&userInfo); err != nil {
		return "", err
	}

	return b.oauthLogin(ctx, &userInfo)
}

func (b *AuthorizationBiz) oauthLogin(ctx context.Context, userInfo auth.IOAuthUser) (string, error) {
	sysUserDo, err := b.oAuthRepo.OAuthUserFirstOrCreate(ctx, userInfo)
	if !types.IsNil(err) {
		return "", err
	}
	if helper.CheckEmail(sysUserDo.Email) != nil {
		authUserDo, err := b.oAuthRepo.GetSysUserByOAuthID(ctx, userInfo.GetOAuthID(), userInfo.GetAPP())
		if err != nil {
			return "", err
		}
		redirect := fmt.Sprintf("%s?oauth_id=%d/#/oauth/register/email", b.redirectURL, authUserDo.ID)
		return redirect, nil
	}

	// 生成token
	base, err := b.getJwtBaseInfo(ctx, sysUserDo, 0)
	if !types.IsNil(err) {
		return "", err
	}

	jwtClaims := middleware.NewJwtClaims(base)
	jwtToken, err := jwtClaims.GetToken()
	if !types.IsNil(err) {
		return "", err
	}
	redirect := fmt.Sprintf("%s?token=%s", b.redirectURL, jwtToken)
	return redirect, nil
}

func (b *AuthorizationBiz) OauthLogin(ctx context.Context, oauthParams *auth.OauthLoginParams) (*bo.RefreshTokenReply, error) {
	if err := b.oAuthRepo.CheckVerifyEmailCode(ctx, oauthParams.Email, oauthParams.Code); err != nil {
		return nil, err
	}

	sysUserDo, err := b.oAuthRepo.SetEmail(ctx, oauthParams.OAuthID, oauthParams.Email)
	if err != nil {
		return nil, err
	}
	// 生成token
	base, err := b.getJwtBaseInfo(ctx, sysUserDo, 0)
	if !types.IsNil(err) {
		return nil, err
	}

	jwtClaims := middleware.NewJwtClaims(base)
	return &bo.RefreshTokenReply{
		User:      sysUserDo,
		JwtClaims: jwtClaims,
	}, nil
}

// OAuthLoginVerifyEmail 验证邮箱
func (b *AuthorizationBiz) OAuthLoginVerifyEmail(ctx context.Context, e string) error {
	if err := helper.CheckEmail(e); err != nil {
		return err
	}
	return b.oAuthRepo.SendVerifyEmail(ctx, e)
}
