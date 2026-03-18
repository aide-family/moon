// Package senderv1 provides domain factories for the rabbit sender service.
package senderv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = NewRegistry()

func NewRegistry() Registry {
	return &registry{
		senderV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]SenderFactoryV1)),
	}
}

type (
	Registry interface {
		RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1)
		GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool)
	}

	SenderFactoryV1 func(c *config.DomainConfig) (apiv1.SenderServer, func() error, error)
)

type registry struct {
	senderV1 *safety.SyncMap[config.DomainConfig_Driver, SenderFactoryV1]
}

func (r *registry) RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1) {
	r.senderV1.Set(name, factory)
}

func (r *registry) GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool) {
	return r.senderV1.Get(name)
}

func RegisterSenderV1Factory(name config.DomainConfig_Driver, factory SenderFactoryV1) {
	globalRegistry.RegisterSenderV1Factory(name, factory)
}

func GetSenderV1Factory(name config.DomainConfig_Driver) (SenderFactoryV1, bool) {
	return globalRegistry.GetSenderV1Factory(name)
}

