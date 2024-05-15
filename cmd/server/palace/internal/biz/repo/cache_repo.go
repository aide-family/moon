package repo

import (
	"github.com/aide-cloud/moon/pkg/conn"
)

type CacheRepo interface {
	Cacher() conn.Cache
}
