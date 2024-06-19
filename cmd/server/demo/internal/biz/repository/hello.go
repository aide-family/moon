package repository

import (
	"context"
)

// Hello .
type Hello interface {
	// SayHello .
	SayHello(ctx context.Context, name string) (string, error)
}
