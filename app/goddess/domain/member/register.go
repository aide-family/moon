// Package member provides domain factory registration for the member service.
package member

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	"google.golang.org/protobuf/types/known/anypb"

	v1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new member registry.
func newRegistry() *registry {
	return &registry{
		memberV1: domainregister.NewRegistry[MemberFactoryV1](),
	}
}

// MemberFactoryV1 is the factory function for the member service.
type MemberFactoryV1 func(c *config.DomainConfig, driver *anypb.Any) (v1.MemberServer, func() error, error)

type registry struct {
	memberV1 *domainregister.Registry[MemberFactoryV1]
}

// RegisterMemberV1Factory registers a new member factory.
func RegisterMemberV1Factory(name config.DomainConfig_Driver, factory MemberFactoryV1) {
	globalRegistry.memberV1.Register(name, factory)
}

// GetMemberV1Factory gets a member factory.
// If the member factory is not found, it will return false.
// If the member factory is found, it will return true and the member factory.
func GetMemberV1Factory(name config.DomainConfig_Driver) (MemberFactoryV1, bool) {
	return globalRegistry.memberV1.Get(name)
}
