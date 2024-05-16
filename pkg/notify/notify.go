package notify

import (
	"context"
)

type (
	Msg    = map[string]any
	Notify interface {
		Send(ctx context.Context, msg Msg) error
	}
)
