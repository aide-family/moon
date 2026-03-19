// Package webhook provides domain factory registration for the webhook service.
package webhook

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new webhook registry.
func newRegistry() *registry {
	return &registry{
		webhookV1: domainregister.NewRegistry[WebhookFactoryV1](),
	}
}

// WebhookFactoryV1 is the factory function for the webhook service.
type WebhookFactoryV1 func(c *config.DomainConfig) (apiv1.WebhookServer, func() error, error)

type registry struct {
	webhookV1 *domainregister.Registry[WebhookFactoryV1]
}

// RegisterWebhookV1Factory registers a new webhook factory.
func RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1) {
	globalRegistry.webhookV1.Register(name, factory)
}

// GetWebhookV1Factory gets a webhook factory.
// If the webhook factory is not found, it will return false.
// If the webhook factory is found, it will return true and the webhook factory.
func GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool) {
	return globalRegistry.webhookV1.Get(name)
}
