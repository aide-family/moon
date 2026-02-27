// Package authv1 is the repository for the auth service.
package authv1

import (
	"context"

	"github.com/aide-family/magicbox/oauth"
)

type Repository interface {
	Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error)
}
