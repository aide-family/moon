package rediscacher

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/redis/go-redis/v9"
)

// NewRedisCacher create redis cacher
func NewRedisCacher(cli *redis.Client) conn.Cache {
	return &redisCacher{
		cli: cli,
	}
}

type redisCacher struct {
	cli *redis.Client
}

func (l *redisCacher) Get(ctx context.Context, key string) (string, error) {
	return l.cli.Get(ctx, key).Result()
}

func (l *redisCacher) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return l.cli.Set(ctx, key, value, expiration).Err()
}

func (l *redisCacher) Delete(ctx context.Context, key string) error {
	return l.cli.Del(ctx, key).Err()
}

func (l *redisCacher) Close() error {
	return l.cli.Close()
}

func (l *redisCacher) Exist(ctx context.Context, key string) bool {
	return l.cli.Exists(ctx, key).Val() > 0
}
