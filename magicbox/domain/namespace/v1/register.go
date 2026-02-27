// Package namespacev1 is the namespace service implementation.
package namespacev1

import (
	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new namespace registry.
func NewRegistry() Registry {
	return &registry{
		namespaceV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]NamespaceFactoryV1)),
	}
}

// NamespaceFactoryV1 is the factory function for the namespace service.
type NamespaceFactoryV1 func(c *config.DomainConfig) (apiv1.NamespaceServer, func() error, error)

type Registry interface {
	RegisterNamespaceV1Factory(name config.DomainConfig_Driver, factory NamespaceFactoryV1)
	GetNamespaceV1Factory(name config.DomainConfig_Driver) (NamespaceFactoryV1, bool)
}

type registry struct {
	namespaceV1 *safety.SyncMap[config.DomainConfig_Driver, NamespaceFactoryV1]
}

func (r *registry) RegisterNamespaceV1Factory(name config.DomainConfig_Driver, factory NamespaceFactoryV1) {
	r.namespaceV1.Set(name, factory)
}

func (r *registry) GetNamespaceV1Factory(name config.DomainConfig_Driver) (NamespaceFactoryV1, bool) {
	return r.namespaceV1.Get(name)
}

// RegisterNamespaceV1Factory registers a new namespace factory.
func RegisterNamespaceV1Factory(name config.DomainConfig_Driver, factory NamespaceFactoryV1) {
	globalRegistry.RegisterNamespaceV1Factory(name, factory)
}

// GetNamespaceV1Factory gets a namespace factory.
// If the namespace factory is not found, it will return false.
// If the namespace factory is found, it will return true and the namespace factory.
func GetNamespaceV1Factory(name config.DomainConfig_Driver) (NamespaceFactoryV1, bool) {
	return globalRegistry.GetNamespaceV1Factory(name)
}
