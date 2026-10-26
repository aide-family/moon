package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// OAuthRepo 按第三方身份查找或绑定本地用户。
type OAuthRepo interface {
	OAuthAccount(ctx context.Context, provider, openID string) (*bo.Account, error)
	LinkOAuth(ctx context.Context, userUID, provider, openID, email string) error
	// MoveOAuth 把一条第三方绑定从 fromUserUID 改绑到 toUserUID，并同步绑定邮箱。
	// 绑定不存在时返回 NotFound。
	MoveOAuth(ctx context.Context, provider, openID, fromUserUID, toUserUID, email string) error
}
