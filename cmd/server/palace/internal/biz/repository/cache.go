package repository

import (
	"github.com/aide-family/moon/pkg/util/conn"
)

type Cache interface {
	Cacher() conn.Cache
}
