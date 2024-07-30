package repository

import (
	"github.com/aide-family/moon/pkg/util/conn"
)

// CacheRepo cache repo
type CacheRepo interface {
	Cacher() conn.Cache
}
