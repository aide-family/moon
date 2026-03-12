package repository

import (
	"context"
)

// Transaction runs fn in a single database transaction. If ctx already has a transaction, fn reuses it.
type Transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
