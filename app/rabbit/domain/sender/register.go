// Package sender provides domain factory registration for the sender service.
package sender

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new sender registry.
func NewRegistry() Registry {
	return &registry{
		senderV1: domainregister.NewRegistry[SenderFactoryV1](),
	}
}

// SenderFactoryV1 is the factory function for the sender service.
type SenderFactoryV1 func(c *config.DomainConfig) (apiv1.SenderServer, func() error, error)

type Registry interface {
	RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1)
	GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool)
}

type registry struct {
	senderV1 *domainregister.Registry[SenderFactoryV1]
}

func (r *registry) RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1) {
	r.senderV1.Register(name, factory)
}

func (r *registry) GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool) {
	return r.senderV1.Get(name)
}

// RegisterSenderV1Factory registers a new sender factory.
func RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1) {
	globalRegistry.RegisterSenderV1Factory(name, factory)
}

func GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool) {
	return globalRegistry.GetSenderV1Factory(name)
}

