package repo

import (
	"github.com/aide-family/moon/pkg/conn"
)

type CacheRepo interface {
	Cacher() conn.Cache
}
