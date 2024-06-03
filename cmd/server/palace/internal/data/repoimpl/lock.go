package repoimpl

import (
	"context"
	"time"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
)

func NewLockRepository(data *data.Data) repository.Lock {
	return &lockRepositoryImpl{data: data}
}

type lockRepositoryImpl struct {
	data *data.Data
}

func (l *lockRepositoryImpl) Lock(ctx context.Context, key string, expire time.Duration) error {
	// 判断是否存在
	if l.data.GetCacher().Exist(ctx, key) {
		return bo.LockFailedErr(ctx)
	}
	return l.data.GetCacher().Set(ctx, key, key, expire)
}

func (l *lockRepositoryImpl) UnLock(ctx context.Context, key string) error {
	return l.data.GetCacher().Delete(ctx, key)
}
