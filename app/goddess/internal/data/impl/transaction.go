package impl

import (
	"context"

	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/data"
)

func NewTransaction(d *data.Data) repository.Transaction {
	return NewTransactionWithDB(d.DB())
}

func NewTransactionWithDB(db *gorm.DB) repository.Transaction {
	return &transaction{db: db}
}

type transaction struct {
	db *gorm.DB
}

func (t *transaction) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, ok := GetTransaction(ctx)
	if ok {
		return fn(WithTransaction(ctx, tx))
	}
	return t.db.Transaction(func(tx *gorm.DB) error {
		return fn(WithTransaction(ctx, tx))
	})
}

type transactionContext struct{}

func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionContext{}, tx)
}

func GetTransaction(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(transactionContext{}).(*gorm.DB)
	return tx, ok
}
