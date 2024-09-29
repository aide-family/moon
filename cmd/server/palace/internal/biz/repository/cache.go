package repository

import (
	"github.com/aide-family/moon/pkg/plugin/cache"
)

// Cache 缓存接口
type Cache interface {
	// Cacher 获取缓存实例
	Cacher() cache.ICacher
}
