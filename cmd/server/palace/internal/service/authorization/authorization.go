package authorization

import (
	"context"
	_ "embed"
	nhttp "net/http"

	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo/auth"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/captcha"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/oauth2"
)

// Service 权限服务
type Service struct {
	authorizationapi.UnimplementedAuthorizationServer

	captchaBiz       *biz.CaptchaBiz
	authorizationBiz *biz.AuthorizationBiz
}

// NewAuthorizationService 创建权限服务
func NewAuthorizationService(
	captchaBiz *biz.CaptchaBiz,
	authorizationBiz *biz.AuthorizationBiz,
) *Service {
	return &Service{
		captchaBiz:       captchaBiz,
		authorizationBiz: authorizationBiz,
	}
}

// Login 登录
func (s *Service) Login(ctx context.Context, req *authorizationapi.LoginRequest) (*authorizationapi.LoginReply, error) {
	captchaInfo := req.GetCaptcha()
	//// 校验验证码
	if err := s.captchaBiz.VerifyCaptcha(ctx, &bo.ValidateCaptchaParams{
		ID:    captchaInfo.GetId(),
		Value: captchaInfo.GetCode(),
	}); !types.IsNil(err) {
		return nil, err
	}

	params := &bo.LoginParams{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Team:     req.GetTeamID(),
	}
	// 执行登录逻辑
	loginJwtClaims, err := s.authorizationBiz.Login(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}

	token, err := loginJwtClaims.JwtClaims.GetToken()
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.LoginReply{
		User:     builder.NewParamsBuild().WithContext(ctx).UserModuleBuilder().DoUserBuilder().ToAPI(loginJwtClaims.User),
		Token:    token,
		Redirect: "/",
	}, nil
}

// Logout 登出
func (s *Service) Logout(ctx context.Context, _ *authorizationapi.LogoutRequest) (*authorizationapi.LogoutReply, error) {
	jwtClaims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}

	if err := s.authorizationBiz.Logout(ctx, &bo.LogoutParams{
		JwtClaims: jwtClaims,
	}); !types.IsNil(err) {
		return nil, merr.ErrorNotification("系统错误")
	}

	return &authorizationapi.LogoutReply{
		Redirect: "/login",
	}, nil
}

// RefreshToken 刷新token
func (s *Service) RefreshToken(ctx context.Context, req *authorizationapi.RefreshTokenRequest) (*authorizationapi.RefreshTokenReply, error) {
	jwtClaims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	tokenRes, err := s.authorizationBiz.RefreshToken(ctx, &bo.RefreshTokenParams{
		JwtClaims: jwtClaims,
		Team:      req.GetTeamID(),
	})
	if !types.IsNil(err) {
		return nil, err
	}

	token, err := tokenRes.JwtClaims.GetToken()
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.RefreshTokenReply{
		Token: token,
		User:  builder.NewParamsBuild().WithContext(ctx).UserModuleBuilder().DoUserBuilder().ToAPI(tokenRes.User),
	}, nil
}

// Captcha 获取验证码
func (s *Service) Captcha(ctx context.Context, req *authorizationapi.CaptchaReq) (*authorizationapi.CaptchaReply, error) {
	generateCaptcha, err := s.captchaBiz.GenerateCaptcha(ctx, &bo.GenerateCaptchaParams{
		Type:  captcha.Type(req.GetCaptchaType()),
		Theme: captcha.Theme(req.GetTheme()),
		Size:  []int{int(req.GetHeight()), int(req.GetWidth())},
	})
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.CaptchaReply{
		Captcha:     generateCaptcha.Base64s,
		CaptchaType: req.GetCaptchaType(),
		Id:          generateCaptcha.ID,
	}, nil
}

// CheckPermission 权限检查
func (s *Service) CheckPermission(ctx context.Context, req *authorizationapi.CheckPermissionRequest) (*authorizationapi.CheckPermissionReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	if middleware.GetUserRole(ctx).IsAdminOrSuperAdmin() {
		return &authorizationapi.CheckPermissionReply{HasPermission: true}, nil
	}
	teamMemberDo, err := s.authorizationBiz.CheckPermission(ctx, &bo.CheckPermissionParams{
		JwtClaims: claims,
		Operation: req.GetOperation(),
	})
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.CheckPermissionReply{
		HasPermission: true,
		TeamMember:    builder.NewParamsBuild().WithContext(ctx).TeamMemberModuleBuilder().DoTeamMemberBuilder().ToAPI(teamMemberDo),
	}, nil
}

// CheckToken 检查token
func (s *Service) CheckToken(ctx context.Context, _ *authorizationapi.CheckTokenRequest) (*authorizationapi.CheckTokenReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	userDo, err := s.authorizationBiz.CheckToken(ctx, &bo.CheckTokenParams{JwtClaims: claims})
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.CheckTokenReply{
		IsLogin: true,
		User:    builder.NewParamsBuild().WithContext(ctx).UserModuleBuilder().DoUserBuilder().ToAPI(userDo),
	}, nil
}

// OAuthLogin oauth登录
func (s *Service) OAuthLogin(app vobj.OAuthAPP) http.HandlerFunc {
	return func(ctx http.Context) error {
		oauthConf := s.authorizationBiz.GetOAuthConf(app)
		// 重定向到指定地址
		url := oauthConf.AuthCodeURL("state", oauth2.AccessTypeOnline)
		req := ctx.Request()
		resp := ctx.Response()
		resp.Header().Set("Location", url)
		resp.WriteHeader(nhttp.StatusTemporaryRedirect)
		ctx.Reset(resp, req)
		return nil
	}
}

// OAuthLoginCallback oauth登录回调
func (s *Service) OAuthLoginCallback(app vobj.OAuthAPP) http.HandlerFunc {
	return func(ctx http.Context) error {
		code := ctx.Query().Get("code")
		loginRedirect, err := s.authorizationBiz.OAuthLogin(ctx, app, code)
		if err != nil {
			return err
		}
		// 重定向到指定地址
		req := ctx.Request()
		resp := ctx.Response()

		resp.Header().Set("Location", loginRedirect)
		resp.WriteHeader(nhttp.StatusTemporaryRedirect)
		ctx.Reset(resp, req)
		return nil
	}
}

// SetEmailWithLogin 设置邮箱并登录
func (s *Service) SetEmailWithLogin(ctx context.Context, req *authorizationapi.SetEmailWithLoginRequest) (*authorizationapi.SetEmailWithLoginReply, error) {
	// TODO 验证临时密码
	params := &auth.OauthLoginParams{
		Code:    req.GetCode(),
		Email:   req.GetEmail(),
		OAuthID: req.GetOauthID(),
	}
	loginReply, err := s.authorizationBiz.OauthLogin(ctx, params)
	if err != nil {
		return nil, err
	}
	token, err := loginReply.JwtClaims.GetToken()
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.SetEmailWithLoginReply{
		User:  builder.NewParamsBuild().WithContext(ctx).UserModuleBuilder().DoUserBuilder().ToAPI(loginReply.User),
		Token: token,
	}, nil
}

// VerifyEmail 验证邮箱
func (s *Service) VerifyEmail(ctx context.Context, req *authorizationapi.VerifyEmailRequest) (*authorizationapi.VerifyEmailReply, error) {
	// TODO 验证临时密码
	captchaInfo := req.GetCaptcha()
	// 校验验证码
	if err := s.captchaBiz.VerifyCaptcha(ctx, &bo.ValidateCaptchaParams{ID: captchaInfo.GetId(), Value: captchaInfo.GetCode()}); !types.IsNil(err) {
		return nil, err
	}
	if err := s.authorizationBiz.OAuthLoginVerifyEmail(ctx, req.GetEmail()); !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.VerifyEmailReply{}, nil
}
