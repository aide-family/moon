package repository

import (
	"context"

	"github.com/aide-family/magicbox/jwt"

	"github.com/aide-family/goddess/internal/biz/bo"
)

type LoginRepository interface {
	Login(ctx context.Context, req *bo.OAuth2LoginBo) (string, error)
	RefreshToken(ctx context.Context, baseInfo jwt.BaseInfo) (string, error)
}
