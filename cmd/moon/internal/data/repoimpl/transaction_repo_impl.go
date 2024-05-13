package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/moon/internal/data"
	"github.com/aide-cloud/moon/pkg/conn"
	"gorm.io/gorm"
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

func (l *transactionRepoImpl) BizTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	return l.data.GetBizDB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, conn.GormContextTxKey{}, tx)
		return f(txCtx)
	})
}

func (l *transactionRepoImpl) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	return l.data.GetMainDB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, conn.GormContextTxKey{}, tx)
		return f(txCtx)
	})
}
