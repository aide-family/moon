// Package templatev1 provides domain factories for the rabbit template service.
package templatev1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = NewRegistry()

func NewRegistry() Registry {
	return &registry{
		templateV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]TemplateFactoryV1)),
	}
}

type (
	Registry interface {
		RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1)
		GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool)
	}

	TemplateFactoryV1 func(c *config.DomainConfig) (apiv1.TemplateServer, func() error, error)
)

type registry struct {
	templateV1 *safety.SyncMap[config.DomainConfig_Driver, TemplateFactoryV1]
}

func (r *registry) RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1) {
	r.templateV1.Set(name, factory)
}

func (r *registry) GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool) {
	return r.templateV1.Get(name)
}

func RegisterTemplateV1Factory(name config.DomainConfig_Driver, factory TemplateFactoryV1) {
	globalRegistry.RegisterTemplateV1Factory(name, factory)
}

func GetTemplateV1Factory(name config.DomainConfig_Driver) (TemplateFactoryV1, bool) {
	return globalRegistry.GetTemplateV1Factory(name)
}

