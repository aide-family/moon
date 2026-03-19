// Package domain provides reusable domain-related utilities.
package domain

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

// Registry is a thread-safe factory registry keyed by driver.
//
// It is designed to be reused by multiple domain versions (v1/v2/v3).
// Each version instantiates Registry with its own factory type.
type Registry[F any] struct {
	factories *safety.SyncMap[config.DomainConfig_Driver, F]
}

func NewRegistry[F any]() *Registry[F] {
	return &Registry[F]{
		factories: safety.NewSyncMap(make(map[config.DomainConfig_Driver]F)),
	}
}

func (r *Registry[F]) Register(name config.DomainConfig_Driver, factory F) {
	r.factories.Set(name, factory)
}

func (r *Registry[F]) Get(name config.DomainConfig_Driver) (F, bool) {
	return r.factories.Get(name)
}

