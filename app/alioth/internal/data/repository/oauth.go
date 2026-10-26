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
	// ListOAuthByUser 列出账号名下的全部第三方绑定，按绑定时间倒序。
	// 只按 user_uid 过滤，个人资料页据此渲染已绑定的来源。
	ListOAuthByUser(ctx context.Context, userUID string) ([]bo.OAuthLink, error)
	// OAuthByUserProvider 取账号在某个来源下最近绑定的一条身份。
	// 账号在该来源下没有绑定时返回 NotFound。
	OAuthByUserProvider(ctx context.Context, userUID, provider string) (*bo.OAuthLink, error)
	// UnlinkOAuth 删除账号名下的指定绑定。
	// uid 必须和 user_uid 同时匹配，避免按 uid 删到其他账号的绑定；0 行返回 NotFound。
	UnlinkOAuth(ctx context.Context, userUID, uid string) error
}
