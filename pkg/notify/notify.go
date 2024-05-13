package notify

import (
	"context"
)

type (
	Notify interface {
		Send(ctx context.Context, msg string) error
	}
)
