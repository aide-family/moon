package repo

import (
	"github.com/aide-family/moon/pkg/util/conn"
)

// CacheRepo 缓存仓库
type CacheRepo interface {
	// Cacher 获取缓存实例
	Cacher() conn.Cache
}
