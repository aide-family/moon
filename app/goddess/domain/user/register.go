// Package user provides domain factory registration for the user service.
package user

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"

	v1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new user registry.
func newRegistry() *registry {
	return &registry{
		userV1: domainregister.NewRegistry[UserFactoryV1](),
	}
}

// UserFactoryV1 is the factory function for the user service.
type UserFactoryV1 func(c *config.DomainConfig) (v1.UserServer, func() error, error)

type registry struct {
	userV1 *domainregister.Registry[UserFactoryV1]
}

// RegisterUserFactoryV1 registers a new user factory.
func RegisterUserFactoryV1(name config.DomainConfig_Driver, factory UserFactoryV1) {
	globalRegistry.userV1.Register(name, factory)
}

// GetUserFactoryV1 gets a user factory.
// If the user factory is not found, it will return false.
// If the user factory is found, it will return true and the user factory.
func GetUserFactoryV1(name config.DomainConfig_Driver) (UserFactoryV1, bool) {
	return globalRegistry.userV1.Get(name)
}
