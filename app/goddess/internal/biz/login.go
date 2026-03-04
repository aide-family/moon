package biz

import (
	"context"

	"github.com/aide-family/magicbox/jwt"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

func NewLoginBiz(transaction repository.Transaction, authRepo repository.LoginRepository) *LoginBiz {
	return &LoginBiz{transaction: transaction, authRepo: authRepo}
}

type LoginBiz struct {
	transaction repository.Transaction
	authRepo    repository.LoginRepository
}

func (b *LoginBiz) Login(ctx context.Context, req *bo.OAuth2LoginBo) (string, error) {
	var token string
	var err error
	err = b.transaction.Transaction(ctx, func(ctx context.Context) error {
		token, err = b.authRepo.LoginByOAuth2(ctx, req)
		return err
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (b *LoginBiz) RefreshToken(ctx context.Context, baseInfo jwt.BaseInfo) (string, error) {
	return b.authRepo.RefreshToken(ctx, baseInfo)
}

func (b *LoginBiz) EmailLogin(ctx context.Context, email string) (string, error) {
	var token string
	var err error
	err = b.transaction.Transaction(ctx, func(ctx context.Context) error {
		token, err = b.authRepo.LoginByEmail(ctx, email)
		return err
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
