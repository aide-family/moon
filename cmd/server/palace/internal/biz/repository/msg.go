package repository

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
)

// Msg .
type Msg interface {
	// Send 发送消息
	Send(ctx context.Context, msg *bo.Message) error
}
