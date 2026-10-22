package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type CaptchaService struct {
	v1.UnimplementedCaptchaServiceServer
	authenticator *auth.Authenticator
}

func NewCaptchaService(authenticator *auth.Authenticator) *CaptchaService {
	return &CaptchaService{authenticator: authenticator}
}

func (s *CaptchaService) GetCaptcha(ctx context.Context, _ *v1.GetCaptchaRequest) (*v1.GetCaptchaResponse, error) {
	id, image, err := s.authenticator.GetCaptcha(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetCaptchaResponse{Id: id, Image: image}, nil
}

type AuthService struct {
	v1.UnimplementedAuthServiceServer
	authenticator *auth.Authenticator
}

func NewAuthService(authenticator *auth.Authenticator) *AuthService {
	return &AuthService{authenticator: authenticator}
}

func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	session, err := s.authenticator.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetName(), req.GetCaptchaId(), req.GetCaptchaCode())
	if err != nil {
		return nil, err
	}
	return &v1.RegisterResponse{Session: tokenPair(session)}, nil
}

func (s *AuthService) PasswordLogin(ctx context.Context, req *v1.PasswordLoginRequest) (*v1.PasswordLoginResponse, error) {
	session, err := s.authenticator.PasswordLogin(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &v1.PasswordLoginResponse{Session: tokenPair(session)}, nil
}

func (s *AuthService) SendEmailLoginCode(ctx context.Context, req *v1.SendEmailLoginCodeRequest) (*v1.SendEmailLoginCodeResponse, error) {
	if err := s.authenticator.SendEmailLoginCode(ctx, req.GetEmail()); err != nil {
		return nil, err
	}
	return &v1.SendEmailLoginCodeResponse{}, nil
}

func (s *AuthService) EmailLogin(ctx context.Context, req *v1.EmailLoginRequest) (*v1.EmailLoginResponse, error) {
	session, err := s.authenticator.EmailLogin(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		return nil, err
	}
	return &v1.EmailLoginResponse{Session: tokenPair(session)}, nil
}

func (s *AuthService) OAuth2Login(ctx context.Context, req *v1.OAuth2LoginRequest) (*v1.OAuth2LoginResponse, error) {
	session, err := s.authenticator.OAuthLogin(ctx, req.GetProvider(), req.GetCode(), req.GetRedirectUri())
	if err != nil {
		return nil, err
	}
	return &v1.OAuth2LoginResponse{Session: tokenPair(session)}, nil
}

func tokenPair(session *bo.Session) *v1.TokenPair {
	return &v1.TokenPair{AccessToken: session.AccessToken, RefreshToken: session.RefreshToken, User: userItem(session.User)}
}
