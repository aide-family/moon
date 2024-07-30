package notify

import (
	"context"
)

type (
	// Msg is a map of string to any.
	Msg = map[string]any

	// Notify is a notification service.
	Notify interface {
		// Send sends a notification.
		Send(ctx context.Context, msg Msg) error
	}
)
