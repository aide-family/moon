// Package namespace provides domain factory registration for the namespace service.
package namespace

import (
	"github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new namespace registry.
func NewRegistry() Registry {
	return &registry{
		namespaceV1: domainregister.NewRegistry[NamespaceFactoryV1](),
	}
}

// NamespaceFactoryV1 is the factory function for the namespace service.
type NamespaceFactoryV1 func(c *config.DomainConfig) (v1.NamespaceServer, func() error, error)

type Registry interface {
	RegisterNamespaceV1Factory(name config.DomainConfig_Driver, factory NamespaceFactoryV1)
	GetNamespaceV1Factory(name config.DomainConfig_Driver) (NamespaceFactoryV1, bool)
}

type registry struct {
	namespaceV1 *domainregister.Registry[NamespaceFactoryV1]
}

func (r *registry) RegisterNamespaceV1Factory(name config.DomainConfig_Driver, factory NamespaceFactoryV1) {
	r.namespaceV1.Register(name, factory)
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

