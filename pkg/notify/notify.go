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
		// Type returns the type of the notification service.
		Type() string
		// Hash returns the hash of the notification service.
		Hash() string
	}
)
