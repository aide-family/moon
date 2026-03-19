// Package template provides domain factory registration for the template service.
package template

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new template registry.
func newRegistry() *registry {
	return &registry{
		templateV1: domainregister.NewRegistry[TemplateFactoryV1](),
	}
}

// TemplateFactoryV1 is the factory function for the template service.
type TemplateFactoryV1 func(c *config.DomainConfig) (apiv1.TemplateServer, func() error, error)

type registry struct {
	templateV1 *domainregister.Registry[TemplateFactoryV1]
}

// RegisterTemplateV1Factory registers a new template factory.
func RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1) {
	globalRegistry.templateV1.Register(name, factory)
}

// GetTemplateV1Factory gets a template factory.
// If the template factory is not found, it will return false.
// If the template factory is found, it will return true and the template factory.
func GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool) {
	return globalRegistry.templateV1.Get(name)
}
