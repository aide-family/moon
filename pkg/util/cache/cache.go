package cache

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/consts"
)

type Cache interface {
	Delete(key string)
	Store(key string, value string)
	Range(f func(key, value string) bool)
}

var _ Cache = (*redisCache)(nil)

type redisCache struct {
	cache  GlobalCache
	prefix consts.RedisKey
}

func (l *redisCache) Delete(key string) {
	_ = l.cache.HDel(context.Background(), l.prefix.String(), key)
}

func (l *redisCache) Store(key string, value string) {
	args := [][]byte{[]byte(key), []byte(value)}
	_ = l.cache.HSet(context.Background(), l.prefix.String(), args...)
}

func (l *redisCache) Range(f func(key string, value string) bool) {
	resMap, err := l.cache.HGetAll(context.Background(), l.prefix.String())
	if err != nil {
		return
	}
	for k, v := range resMap {
		if !f(k, string(v)) {
			continue
		}
	}
}

func NewRedisCache(cache GlobalCache, prefix consts.RedisKey) Cache {
	return &redisCache{
		cache:  cache,
		prefix: prefix,
	}
}
