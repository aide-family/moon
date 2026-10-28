package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// MessageRepo 读写平台系统消息。
// 未注册用户的消息先按 email 匹配，注册后按 user_uid 或 email 都能看到。
type MessageRepo interface {
	ListMessages(ctx context.Context, userUID, email string, page, pageSize int, unreadOnly bool) ([]bo.SystemMessage, int, error)
	UnreadCount(ctx context.Context, userUID, email string) (int, error)
	MarkRead(ctx context.Context, userUID, email, uid string) error
	MarkAllRead(ctx context.Context, userUID, email string) error
}
