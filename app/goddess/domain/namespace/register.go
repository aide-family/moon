// Package namespace provides domain factory registration for the namespace service.
package namespace

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	"google.golang.org/protobuf/types/known/anypb"

	v1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new namespace registry.
func newRegistry() *registry {
	return &registry{
		namespaceV1: domainregister.NewRegistry[NamespaceFactoryV1](),
	}
}

// NamespaceFactoryV1 is the factory function for the namespace service.
type NamespaceFactoryV1 func(c *config.DomainConfig, driver *anypb.Any) (v1.NamespaceServer, func() error, error)

type registry struct {
	namespaceV1 *domainregister.Registry[NamespaceFactoryV1]
}

// RegisterNamespaceV1Factory registers a new namespace factory.
func RegisterNamespaceV1Factory(name config.DomainConfig_Driver, factory NamespaceFactoryV1) {
	globalRegistry.namespaceV1.Register(name, factory)
}

// GetNamespaceV1Factory gets a namespace factory.
// If the namespace factory is not found, it will return false.
// If the namespace factory is found, it will return true and the namespace factory.
func GetNamespaceV1Factory(name config.DomainConfig_Driver) (NamespaceFactoryV1, bool) {
	return globalRegistry.namespaceV1.Get(name)
}
