package biz

import (
	"context"

	"github.com/aide-family/magicbox/oauth"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewLoginBiz(authRepo repository.LoginRepository) *LoginBiz {
	return &LoginBiz{authRepo: authRepo}
}

type LoginBiz struct {
	authRepo repository.LoginRepository
}

func (b *LoginBiz) Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
	redirectURL, err := b.authRepo.Login(ctx, req)
	if err != nil {
		return "", err
	}
	return redirectURL, nil
}
