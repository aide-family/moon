package repository

import (
	"github.com/aide-family/moon/pkg/plugin/cache"
)

// CacheRepo cache repo
type CacheRepo interface {
	Cacher() cache.ICacher
}
