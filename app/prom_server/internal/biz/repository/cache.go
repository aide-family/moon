package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type (
	CacheRepo interface {
		Client() *redis.Client
		Set(ctx context.Context, key string, value any, expiration time.Duration) error
	}
)
