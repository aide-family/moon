package service

import (
	"context"

	"github.com/aide-family/magicbox/oauth"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

type AuthService struct {
	goddessv1.AuthServiceServer
	loginBiz *biz.LoginBiz
}

func NewAuthService(loginBiz *biz.LoginBiz) *AuthService {
	return &AuthService{loginBiz: loginBiz}
}

func (s *AuthService) OAuth2Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (*goddessv1.LoginReply, error) {
	loginBo := bo.NewOAuth2LoginBo(req)
	token, err := s.loginBiz.Login(ctx, loginBo)
	if err != nil {
		return nil, err
	}
	return &goddessv1.LoginReply{Token: token}, nil
}
