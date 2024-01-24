package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"prometheus-manager/pkg/helper/consts"
)

type Cache interface {
	Delete(key string)
	Store(key string, value string)
	Range(f func(key, value string) bool)
}

var _ Cache = (*redisCache)(nil)

type redisCache struct {
	cache  *redis.Client
	prefix consts.RedisKey
}

func (l *redisCache) Delete(key string) {
	l.cache.HDel(context.Background(), l.prefix.String(), key)
}

func (l *redisCache) Store(key string, value string) {
	args := []any{key, value}
	l.cache.HSet(context.Background(), l.prefix.String(), args)
}

func (l *redisCache) Range(f func(key string, value string) bool) {
	resMap, err := l.cache.HGetAll(context.Background(), l.prefix.String()).Result()
	if err != nil {
		return
	}
	for k, v := range resMap {
		if !f(k, v) {
			continue
		}
	}
}

func NewRedisCache(cache *redis.Client, prefix consts.RedisKey) Cache {
	return &redisCache{
		cache:  cache,
		prefix: prefix,
	}
}
