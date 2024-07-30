package repository

import (
	"github.com/aide-family/moon/pkg/util/conn"
)

// CacheRepo 换成统一repo
type CacheRepo interface {
	// Cacher 获取缓存实现
	Cacher() conn.Cache
}
