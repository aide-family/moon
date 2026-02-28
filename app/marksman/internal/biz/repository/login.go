package repository

import (
	"context"

	"github.com/aide-family/magicbox/oauth"
)

type LoginRepository interface {
	Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error)
}
