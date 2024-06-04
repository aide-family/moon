package repository

import (
	"github.com/aide-family/moon/pkg/conn"
)

type Cache interface {
	Cacher() conn.Cache
}
