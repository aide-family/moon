package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// OAuthRepo 按第三方身份查找或绑定本地用户。
type OAuthRepo interface {
	OAuthAccount(ctx context.Context, provider, openID string) (*bo.Account, error)
	LinkOAuth(ctx context.Context, userUID, provider, openID, email string) error
}
