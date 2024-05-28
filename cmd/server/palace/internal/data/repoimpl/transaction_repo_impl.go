package repoimpl

import (
	"context"

	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/conn"
)

func NewTransactionRepo(data *data.Data) repo.TransactionRepo {
	return &transactionRepoImpl{
		data: data,
	}
}

type (
	transactionRepoImpl struct {
		data *data.Data
	}
)

func (l *transactionRepoImpl) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	return l.data.GetMainDB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, conn.GormContextTxKey{}, tx)
		return f(txCtx)
	})
}
