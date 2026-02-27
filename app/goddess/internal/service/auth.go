package service

import (
	"context"

	"github.com/aide-family/magicbox/oauth"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
)

type AuthService struct {
	loginBiz *biz.LoginBiz
}

func NewAuthService(loginBiz *biz.LoginBiz) *AuthService {
	return &AuthService{loginBiz: loginBiz}
}

func (s *AuthService) Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
	loginBo := bo.NewOAuth2LoginBo(req)
	return s.loginBiz.Login(ctx, loginBo)
}
