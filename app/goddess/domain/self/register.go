// Package self provides domain factory registration for the self service.
package self

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"

	v1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = newRegistry()

// NewRegistry creates a new self registry.
func newRegistry() *registry {
	return &registry{
		selfV1: domainregister.NewRegistry[SelfFactoryV1](),
	}
}

// SelfFactoryV1 is the factory function for the self service.
type SelfFactoryV1 func(c *config.DomainConfig) (v1.SelfServer, func() error, error)

type registry struct {
	selfV1 *domainregister.Registry[SelfFactoryV1]
}

// RegisterSelfFactoryV1 registers a new self factory.
func RegisterSelfFactoryV1(name config.DomainConfig_Driver, factory SelfFactoryV1) {
	globalRegistry.selfV1.Register(name, factory)
}

// GetSelfFactoryV1 gets a self factory.
// If the self factory is not found, it will return false.
// If the self factory is found, it will return true and the self factory.
func GetSelfFactoryV1(name config.DomainConfig_Driver) (SelfFactoryV1, bool) {
	return globalRegistry.selfV1.Get(name)
}
