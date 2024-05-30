package repository

import (
	"context"
	"time"
)

// Lock .
type Lock interface {
	// Lock 加锁
	Lock(ctx context.Context, key string, expire time.Duration) error
	// UnLock 解锁
	UnLock(ctx context.Context, key string) error
}
