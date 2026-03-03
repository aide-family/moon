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
	return b.authRepo.LoginByOAuth2(ctx, req)
}

func (b *LoginBiz) RefreshToken(ctx context.Context, baseInfo jwt.BaseInfo) (string, error) {
	return b.authRepo.RefreshToken(ctx, baseInfo)
}

func (b *LoginBiz) EmailLogin(ctx context.Context, email string) (string, error) {
	return b.authRepo.LoginByEmail(ctx, email)
}
