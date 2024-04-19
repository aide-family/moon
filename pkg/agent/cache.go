package agent

import (
	"context"
	"errors"
	"sync"
	"time"
)

type (
	Cache interface {
		// Get 获取缓存, 对象类型， 不支持基础类型
		Get(key string, value any) error
		// Set 设置缓存， value为对象类型
		Set(key string, value any, expiration time.Duration) error
		// Delete 删除缓存
		Delete(key string) error
		// SetNX 设置缓存，如果key存在，则不设置
		SetNX(key string, value any, expiration time.Duration) bool
		// Exists 判断缓存是否存在
		Exists(key string) bool
		// WithContext 设置上下文
		WithContext(ctx context.Context) Cache
		// Close 关闭缓存
		Close() error
	}
)

var NoCache = errors.New("no cache")

var globalCache Cache
var globalCacheOnce sync.Once

// SetGlobalCache 设置全局缓存
func SetGlobalCache(cache Cache) {
	globalCacheOnce.Do(func() {
		globalCache = cache
	})
}

// GetGlobalCache 获取全局缓存
func GetGlobalCache() Cache {
	return globalCache
}
