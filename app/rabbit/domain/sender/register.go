// Package sender provides domain factory registration for the sender service.
package sender

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new sender registry.
func newRegistry() *registry {
	return &registry{
		senderV1: domainregister.NewRegistry[SenderFactoryV1](),
	}
}

// SenderFactoryV1 is the factory function for the sender service.
type SenderFactoryV1 func(c *config.DomainConfig) (apiv1.SenderServer, func() error, error)

type registry struct {
	senderV1 *domainregister.Registry[SenderFactoryV1]
}

// RegisterSenderV1Factory registers a new sender factory.
func RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1) {
	globalRegistry.senderV1.Register(name, factory)
}

// GetSenderV1Factory gets a sender factory.
// If the sender factory is not found, it will return false.
// If the sender factory is found, it will return true and the sender factory.
func GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool) {
	return globalRegistry.senderV1.Get(name)
}
