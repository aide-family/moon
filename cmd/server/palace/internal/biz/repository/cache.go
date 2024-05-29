package repository

import (
	"github.com/aide-cloud/moon/pkg/conn"
)

type Cache interface {
	Cacher() conn.Cache
}
