package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type OAuth interface {
	Create(ctx context.Context, user bo.IOAuthUser) (do.UserOAuth, error)
	FindByOpenID(ctx context.Context, openID string, app vobj.OAuthAPP) (do.UserOAuth, error)
	SetUser(ctx context.Context, user do.UserOAuth) (do.UserOAuth, error)
}
