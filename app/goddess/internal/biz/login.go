package biz

import (
	"context"

	"github.com/aide-family/magicbox/jwt"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

func NewLoginBiz(authRepo repository.LoginRepository) *LoginBiz {
	return &LoginBiz{authRepo: authRepo}
}

type LoginBiz struct {
	authRepo repository.LoginRepository
}

func (b *LoginBiz) Login(ctx context.Context, req *bo.OAuth2LoginBo) (string, error) {
	redirectURL, err := b.authRepo.Login(ctx, req)
	if err != nil {
		return "", err
	}
	return redirectURL, nil
}

func (b *LoginBiz) RefreshToken(ctx context.Context, baseInfo jwt.BaseInfo) (string, error) {
	return b.authRepo.RefreshToken(ctx, baseInfo)
}
