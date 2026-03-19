// Package template provides domain factory registration for the template service.
package template

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new template registry.
func NewRegistry() Registry {
	return &registry{
		templateV1: domainregister.NewRegistry[TemplateFactoryV1](),
	}
}

// TemplateFactoryV1 is the factory function for the template service.
type TemplateFactoryV1 func(c *config.DomainConfig) (apiv1.TemplateServer, func() error, error)

type Registry interface {
	RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1)
	GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool)
}

type registry struct {
	templateV1 *domainregister.Registry[TemplateFactoryV1]
}

func (r *registry) RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1) {
	r.templateV1.Register(name, factory)
}

func (r *registry) GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool) {
	return r.templateV1.Get(name)
}

// RegisterTemplateV1Factory registers a new template factory.
func RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1) {
	globalRegistry.RegisterTemplateV1Factory(name, factory)
}

func GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool) {
	return globalRegistry.GetTemplateV1Factory(name)
}

