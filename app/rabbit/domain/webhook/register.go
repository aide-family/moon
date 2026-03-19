// Package webhook provides domain factory registration for the webhook service.
package webhook

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new webhook registry.
func NewRegistry() Registry {
	return &registry{
		webhookV1: domainregister.NewRegistry[WebhookFactoryV1](),
	}
}

// WebhookFactoryV1 is the factory function for the webhook service.
type WebhookFactoryV1 func(c *config.DomainConfig) (apiv1.WebhookServer, func() error, error)

type Registry interface {
	RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1)
	GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool)
}

type registry struct {
	webhookV1 *domainregister.Registry[WebhookFactoryV1]
}

func (r *registry) RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1) {
	r.webhookV1.Register(name, factory)
}

func (r *registry) GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool) {
	return r.webhookV1.Get(name)
}

// RegisterWebhookV1Factory registers a new webhook factory.
func RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1) {
	globalRegistry.RegisterWebhookV1Factory(name, factory)
}

func GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool) {
	return globalRegistry.GetWebhookV1Factory(name)
}

