package repoimpl

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
)

// NewLockRepository 创建全局锁
func NewLockRepository(data *data.Data) repository.Lock {
	return &lockRepositoryImpl{data: data}
}

type lockRepositoryImpl struct {
	data *data.Data
}

func (l *lockRepositoryImpl) Lock(ctx context.Context, key string, expire time.Duration) error {
	locked, err := l.data.GetCacher().Client().SetNX(ctx, key, key, expire).Result()
	if err != nil {
		return err
	}
	// 判断是否存在
	if !locked {
		return merr.ErrorI18nToastDatasourceSyncing(ctx)
	}
	return nil
}

func (l *lockRepositoryImpl) UnLock(ctx context.Context, key string) error {
	return l.data.GetCacher().Client().Del(ctx, key).Err()
}
