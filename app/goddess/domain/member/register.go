// Package member provides domain factory registration for the member service.
package member

import (
	"github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new member registry.
func NewRegistry() Registry {
	return &registry{
		memberV1: domainregister.NewRegistry[MemberFactoryV1](),
	}
}

// MemberFactoryV1 is the factory function for the member service.
type MemberFactoryV1 func(c *config.DomainConfig) (v1.MemberServer, func() error, error)

type Registry interface {
	RegisterMemberV1Factory(name config.DomainConfig_Driver, factory MemberFactoryV1)
	GetMemberV1Factory(name config.DomainConfig_Driver) (MemberFactoryV1, bool)
}

type registry struct {
	memberV1 *domainregister.Registry[MemberFactoryV1]
}

func (r *registry) RegisterMemberV1Factory(name config.DomainConfig_Driver, factory MemberFactoryV1) {
	r.memberV1.Register(name, factory)
}

func (r *registry) GetMemberV1Factory(name config.DomainConfig_Driver) (MemberFactoryV1, bool) {
	return r.memberV1.Get(name)
}

// RegisterMemberV1Factory registers a new member factory.
func RegisterMemberV1Factory(name config.DomainConfig_Driver, factory MemberFactoryV1) {
	globalRegistry.RegisterMemberV1Factory(name, factory)
}

func GetMemberV1Factory(name config.DomainConfig_Driver) (MemberFactoryV1, bool) {
	return globalRegistry.GetMemberV1Factory(name)
}

