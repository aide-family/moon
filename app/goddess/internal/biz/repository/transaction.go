package repository

import (
	"context"
)

type Transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
