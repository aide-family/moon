package authorization

import (
	"context"

	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/captcha"
	"github.com/aide-family/moon/pkg/util/types"
)

type Service struct {
	authorizationapi.UnimplementedAuthorizationServer

	captchaBiz       *biz.CaptchaBiz
	authorizationBiz *biz.AuthorizationBiz
}

func NewAuthorizationService(
	captchaBiz *biz.CaptchaBiz,
	authorizationBiz *biz.AuthorizationBiz,
) *Service {
	return &Service{
		captchaBiz:       captchaBiz,
		authorizationBiz: authorizationBiz,
	}
}

func (s *Service) Login(ctx context.Context, req *authorizationapi.LoginRequest) (*authorizationapi.LoginReply, error) {
	captchaInfo := req.GetCaptcha()
	// 校验验证码
	if err := s.captchaBiz.VerifyCaptcha(ctx, &bo.ValidateCaptchaParams{
		Id:    captchaInfo.GetId(),
		Value: captchaInfo.GetCode(),
	}); !types.IsNil(err) {
		return nil, err
	}

	// 执行登录逻辑
	loginJwtClaims, err := s.authorizationBiz.Login(ctx, &bo.LoginParams{
		Username:   req.GetUsername(),
		EnPassword: req.GetPassword(),
		Team:       req.GetTeamId(),
	})
	if !types.IsNil(err) {
		return nil, err
	}

	token, err := loginJwtClaims.JwtClaims.GetToken()
	if !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.LoginReply{
		User:     build.NewUserBuilder(loginJwtClaims.User).ToApi(),
		Token:    token,
		Redirect: "/",
	}, nil
}

func (s *Service) Logout(ctx context.Context, _ *authorizationapi.LogoutRequest) (*authorizationapi.LogoutReply, error) {
	jwtClaims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
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

func (s *Service) RefreshToken(ctx context.Context, req *authorizationapi.RefreshTokenRequest) (*authorizationapi.RefreshTokenReply, error) {
	jwtClaims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	tokenRes, err := s.authorizationBiz.RefreshToken(ctx, &bo.RefreshTokenParams{
		JwtClaims: jwtClaims,
		Team:      req.GetTeamId(),
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
		User:  build.NewUserBuilder(tokenRes.User).ToApi(),
	}, nil
}

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
		Id:          generateCaptcha.Id,
	}, nil
}

// CheckPermission 权限检查
func (s *Service) CheckPermission(ctx context.Context, req *authorizationapi.CheckPermissionRequest) (*authorizationapi.CheckPermissionReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	if claims.IsAdminRole() {
		return &authorizationapi.CheckPermissionReply{HasPermission: true}, nil
	}
	if err := s.authorizationBiz.CheckPermission(ctx, &bo.CheckPermissionParams{
		JwtClaims: claims,
		Operation: req.GetOperation(),
	}); !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.CheckPermissionReply{HasPermission: true}, nil
}

// CheckToken 检查token
func (s *Service) CheckToken(ctx context.Context, _ *authorizationapi.CheckTokenRequest) (*authorizationapi.CheckTokenReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	if err := s.authorizationBiz.CheckToken(ctx, &bo.CheckTokenParams{JwtClaims: claims}); !types.IsNil(err) {
		return nil, err
	}
	return &authorizationapi.CheckTokenReply{IsLogin: true}, nil
}
