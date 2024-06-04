package repository

import (
	"github.com/aide-family/moon/pkg/conn"
)

type CacheRepo interface {
	Cacher() conn.Cache
}
