package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/plugin/cache"
)

type Cache interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

const (
	EmailSendKey cache.K = "rabbit:email:send"
)
