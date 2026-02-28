package service

import (
	"context"

	"github.com/aide-family/magicbox/oauth"

	"github.com/aide-family/marksman/internal/biz"
)

type AuthService struct {
	loginBiz *biz.LoginBiz
}

func NewAuthService(loginBiz *biz.LoginBiz) *AuthService {
	return &AuthService{loginBiz: loginBiz}
}

func (s *AuthService) Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
	return s.loginBiz.Login(ctx, req)
}
