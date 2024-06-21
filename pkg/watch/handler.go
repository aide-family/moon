package watch

import (
	"context"
)

type (
	// Handler 消息处理
	Handler interface {
		Handle(ctx context.Context, msg *Message, storage Storage) error
	}
)
