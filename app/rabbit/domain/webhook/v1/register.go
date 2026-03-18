// Package webhookv1 provides domain factories for the rabbit webhook service.
package webhookv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = NewRegistry()

func NewRegistry() Registry {
	return &registry{
		webhookV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]WebhookFactoryV1)),
	}
}

type (
	Registry interface {
		RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1)
		GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool)
	}

	WebhookFactoryV1 func(c *config.DomainConfig) (apiv1.WebhookServer, func() error, error)
)

type registry struct {
	webhookV1 *safety.SyncMap[config.DomainConfig_Driver, WebhookFactoryV1]
}

func (r *registry) RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1) {
	r.webhookV1.Set(name, factory)
}

func (r *registry) GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool) {
	return r.webhookV1.Get(name)
}

func RegisterWebhookV1Factory(name config.DomainConfig_Driver, factory WebhookFactoryV1) {
	globalRegistry.RegisterWebhookV1Factory(name, factory)
}

func GetWebhookV1Factory(name config.DomainConfig_Driver) (WebhookFactoryV1, bool) {
	return globalRegistry.GetWebhookV1Factory(name)
}

