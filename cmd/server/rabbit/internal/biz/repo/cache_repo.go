package repo

import (
	"github.com/aide-family/moon/pkg/util/conn"
)

type CacheRepo interface {
	Cacher() conn.Cache
}
