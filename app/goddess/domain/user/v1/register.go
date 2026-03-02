// Package userv1 is the user service implementation.
package userv1

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new user registry.
func NewRegistry() Registry {
	return &registry{
		userV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]UserFactoryV1)),
	}
}

// UserFactoryV1 is the factory function for the user service.
type (
	UserFactoryV1 func(c *config.DomainConfig) (goddessv1.UserServer, func() error, error)
	Registry      interface {
		RegisterUserFactoryV1(name config.DomainConfig_Driver, factory UserFactoryV1)
		GetUserFactoryV1(name config.DomainConfig_Driver) (UserFactoryV1, bool)
	}
)

type registry struct {
	userV1 *safety.SyncMap[config.DomainConfig_Driver, UserFactoryV1]
}

func (r *registry) RegisterUserFactoryV1(name config.DomainConfig_Driver, factory UserFactoryV1) {
	r.userV1.Set(name, factory)
}

func (r *registry) GetUserFactoryV1(name config.DomainConfig_Driver) (UserFactoryV1, bool) {
	return r.userV1.Get(name)
}

// RegisterUserFactoryV1 registers a new user factory.
func RegisterUserFactoryV1(name config.DomainConfig_Driver, factory UserFactoryV1) {
	globalRegistry.RegisterUserFactoryV1(name, factory)
}

// GetUserFactoryV1 gets a user factory.
// If the user factory is not found, it will return false.
// If the user factory is found, it will return true and the user factory.
func GetUserFactoryV1(name config.DomainConfig_Driver) (UserFactoryV1, bool) {
	return globalRegistry.GetUserFactoryV1(name)
}
