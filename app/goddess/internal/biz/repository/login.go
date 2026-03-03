package repository

import (
	"context"

	"github.com/aide-family/magicbox/jwt"

	"github.com/aide-family/goddess/internal/biz/bo"
)

type LoginRepository interface {
	LoginByOAuth2(ctx context.Context, req *bo.OAuth2LoginBo) (string, error)
	LoginByEmail(ctx context.Context, email string) (string, error)
	RefreshToken(ctx context.Context, baseInfo jwt.BaseInfo) (string, error)
}
