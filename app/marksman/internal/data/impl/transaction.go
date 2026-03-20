package impl

import (
	"context"

	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
)

func NewTransactionRepository(d *data.Data) repository.Transaction {
	return NewTransactionRepositoryWithDB(d.DB())
}

func NewTransactionRepositoryWithDB(db *gorm.DB) repository.Transaction {
	return &transactionRepository{db: db}
}

type transactionRepository struct {
	db *gorm.DB
}

func (t *transactionRepository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, ok := GetTransaction(ctx, t.db)
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

func GetTransaction(ctx context.Context, db *gorm.DB) (*gorm.DB, bool) {
	tx, ok := ctx.Value(transactionContext{}).(*gorm.DB)
	if ok {
		return tx, ok
	}
	return db, false
}

func getDBWithTransaction(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx, ok := GetTransaction(ctx, db)
	if ok {
		return tx
	}
	return db
}
