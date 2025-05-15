package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
)

func NewLock(cacheRepo repository.Cache, logger log.Logger) *Lock {
	return &Lock{
		cacheRepo: cacheRepo,
		helper:    log.NewHelper(log.With(logger, "module", "biz.lock")),
	}
}

type Lock struct {
	cacheRepo repository.Cache
	helper    *log.Helper
}

func (l *Lock) LockByAPP(ctx context.Context, requestId string, app vobj.APP) bool {
	key := vobj.SendLockKey.Key(requestId, app.String())
	locked, err := l.cacheRepo.Lock(ctx, key, 2*time.Hour)
	if err != nil {
		l.helper.WithContext(ctx).Warnw("msg", "failed to lock", "key", key, "err", err)
		return false
	}
	return locked
}

func (l *Lock) UnlockByAPP(ctx context.Context, requestId string, app vobj.APP) {
	key := vobj.SendLockKey.Key(requestId, app.String())
	err := l.cacheRepo.Unlock(ctx, key)
	if err != nil {
		l.helper.WithContext(ctx).Warnw("msg", "failed to unlock", "key", key, "err", err)
	}
}
