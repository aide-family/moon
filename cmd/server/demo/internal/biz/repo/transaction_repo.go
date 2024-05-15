package repo

import (
	"context"
)

type TransactionRepo interface {
	Transaction(ctx context.Context, f func(ctx context.Context) error) error

	BizTransaction(ctx context.Context, f func(ctx context.Context) error) error
}
