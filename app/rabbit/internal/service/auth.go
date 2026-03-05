package service

import (
	"context"

	"github.com/aide-family/magicbox/oauth"

	"github.com/aide-family/rabbit/internal/biz"
)

type AuthService struct {
	*biz.LoginBiz
}

func NewAuthService(loginBiz *biz.LoginBiz) *AuthService {
	return &AuthService{LoginBiz: loginBiz}
}

func (s *AuthService) Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
	reply, err := s.LoginBiz.OAuth2Login(ctx, req)
	if err != nil {
		return "", err
	}
	return reply.Token, nil
}
