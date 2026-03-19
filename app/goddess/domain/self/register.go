// Package self provides domain factory registration for the self service.
package self

import (
	"github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new self registry.
func NewRegistry() Registry {
	return &registry{
		selfV1: domainregister.NewRegistry[SelfFactoryV1](),
	}
}

// SelfFactoryV1 is the factory function for the self service.
type SelfFactoryV1 func(c *config.DomainConfig) (v1.SelfServer, func() error, error)

type Registry interface {
	RegisterSelfFactoryV1(name config.DomainConfig_Driver, factory SelfFactoryV1)
	GetSelfFactoryV1(name config.DomainConfig_Driver) (SelfFactoryV1, bool)
}

type registry struct {
	selfV1 *domainregister.Registry[SelfFactoryV1]
}

func (r *registry) RegisterSelfFactoryV1(name config.DomainConfig_Driver, factory SelfFactoryV1) {
	r.selfV1.Register(name, factory)
}

func (r *registry) GetSelfFactoryV1(name config.DomainConfig_Driver) (SelfFactoryV1, bool) {
	return r.selfV1.Get(name)
}

// RegisterSelfFactoryV1 registers a new self factory.
func RegisterSelfFactoryV1(name config.DomainConfig_Driver, factory SelfFactoryV1) {
	globalRegistry.RegisterSelfFactoryV1(name, factory)
}

// GetSelfFactoryV1 gets a self factory.
// If the self factory is not found, it will return false.
// If the self factory is found, it will return true and the self factory.
func GetSelfFactoryV1(name config.DomainConfig_Driver) (SelfFactoryV1, bool) {
	return globalRegistry.GetSelfFactoryV1(name)
}

