package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/auth"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

type CaptchaService struct {
	v1.UnimplementedCaptchaServiceServer
	authenticator *auth.Authenticator
}

func NewCaptchaService(authenticator *auth.Authenticator) *CaptchaService {
	return &CaptchaService{authenticator: authenticator}
}

func (s *CaptchaService) GetCaptcha(ctx context.Context, _ *v1.GetCaptchaRequest) (*v1.GetCaptchaResponse, error) {
	id, image, kind, err := s.authenticator.GetCaptcha(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetCaptchaResponse{Id: id, Image: image, Kind: kind}, nil
}

type AuthService struct {
	v1.UnimplementedAuthServiceServer
	authenticator *auth.Authenticator
}

func NewAuthService(authenticator *auth.Authenticator) *AuthService {
	return &AuthService{authenticator: authenticator}
}

func (s *AuthService) SendRegisterEmailCode(ctx context.Context, req *v1.SendRegisterEmailCodeRequest) (*v1.SendRegisterEmailCodeResponse, error) {
	if err := s.authenticator.SendRegisterEmailCode(ctx, req.GetEmail(), req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
		return nil, err
	}
	return &v1.SendRegisterEmailCodeResponse{}, nil
}

func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	session, err := s.authenticator.Register(ctx, req.GetEmail(), req.GetCode(), req.GetName(), req.GetCaptchaId(), req.GetCaptchaCode())
	if err != nil {
		return nil, err
	}
	return &v1.RegisterResponse{Session: tokenPair(session)}, nil
}

func (s *AuthService) SendEmailLoginCode(ctx context.Context, req *v1.SendEmailLoginCodeRequest) (*v1.SendEmailLoginCodeResponse, error) {
	if err := s.authenticator.SendEmailLoginCode(ctx, req.GetEmail(), req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
		return nil, err
	}
	return &v1.SendEmailLoginCodeResponse{}, nil
}

func (s *AuthService) EmailLogin(ctx context.Context, req *v1.EmailLoginRequest) (*v1.EmailLoginResponse, error) {
	session, err := s.authenticator.EmailLogin(ctx, req.GetEmail(), req.GetCode(), req.GetCaptchaId(), req.GetCaptchaCode())
	if err != nil {
		return nil, err
	}
	return &v1.EmailLoginResponse{Session: tokenPair(session)}, nil
}

func (s *AuthService) ListOAuth2Providers(ctx context.Context, req *v1.ListOAuth2ProvidersRequest) (*v1.ListOAuth2ProvidersResponse, error) {
	providers, err := s.authenticator.OAuthProviders(req.GetRedirectUri(), req.GetState())
	if err != nil {
		return nil, err
	}
	items := make([]*v1.OAuth2Provider, 0, len(providers))
	for _, item := range providers {
		items = append(items, &v1.OAuth2Provider{
			Provider:     item.Provider,
			Name:         item.Name,
			AuthorizeUrl: item.AuthorizeURL,
		})
	}
	return &v1.ListOAuth2ProvidersResponse{Providers: items}, nil
}

func (s *AuthService) OAuth2Login(ctx context.Context, req *v1.OAuth2LoginRequest) (*v1.OAuth2LoginResponse, error) {
	session, binding, err := s.authenticator.OAuthLogin(ctx, req.GetProvider(), req.GetCode(), req.GetRedirectUri())
	if err != nil {
		return nil, err
	}
	if binding != nil {
		// 第三方没有可用邮箱：不下发会话，前端据 bind_token 引导用户补充并验证邮箱。
		return &v1.OAuth2LoginResponse{
			EmailBindingRequired: true,
			BindToken:            binding.Token,
			SuggestedEmail:       binding.SuggestedEmail,
		}, nil
	}
	return &v1.OAuth2LoginResponse{Session: tokenPair(session)}, nil
}

func (s *AuthService) SendOAuthBindEmailCode(ctx context.Context, req *v1.SendOAuthBindEmailCodeRequest) (*v1.SendOAuthBindEmailCodeResponse, error) {
	if err := s.authenticator.SendOAuthBindEmailCode(ctx, req.GetBindToken(), req.GetEmail(), req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
		return nil, err
	}
	return &v1.SendOAuthBindEmailCodeResponse{}, nil
}

func (s *AuthService) BindOAuthEmail(ctx context.Context, req *v1.BindOAuthEmailRequest) (*v1.BindOAuthEmailResponse, error) {
	session, err := s.authenticator.BindOAuthEmail(ctx, req.GetBindToken(), req.GetEmail(), req.GetCode(), req.GetCaptchaId(), req.GetCaptchaCode())
	if err != nil {
		return nil, err
	}
	return &v1.BindOAuthEmailResponse{Session: tokenPair(session)}, nil
}

func tokenPair(session *bo.Session) *v1.TokenPair {
	return &v1.TokenPair{AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, User: userItem(session.User)}
}
