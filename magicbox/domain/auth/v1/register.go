// Package authv1 is the auth service implementation.
package authv1

import (
	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new auth registry.
func NewRegistry() Registry {
	return &registry{
		authV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]AuthFactoryV1)),
		userV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]UserFactoryV1)),
	}
}

// AuthFactoryV1 is the factory function for the auth service.
type AuthFactoryV1 func(c *config.DomainConfig, jwtConfig *config.JWT) (Repository, func() error, error)

// UserFactoryV1 is the factory function for the user service.
type UserFactoryV1 func(c *config.DomainConfig) (apiv1.UserServer, func() error, error)
type Registry interface {
	RegisterAuthV1Factory(name config.DomainConfig_Driver, factory AuthFactoryV1)
	GetAuthV1Factory(name config.DomainConfig_Driver) (AuthFactoryV1, bool)
	RegisterUserFactoryV1(name config.DomainConfig_Driver, factory UserFactoryV1)
	GetUserFactoryV1(name config.DomainConfig_Driver) (UserFactoryV1, bool)
}

type registry struct {
	authV1 *safety.SyncMap[config.DomainConfig_Driver, AuthFactoryV1]
	userV1 *safety.SyncMap[config.DomainConfig_Driver, UserFactoryV1]
}

func (r *registry) RegisterAuthV1Factory(name config.DomainConfig_Driver, factory AuthFactoryV1) {
	r.authV1.Set(name, factory)
}

func (r *registry) GetAuthV1Factory(name config.DomainConfig_Driver) (AuthFactoryV1, bool) {
	return r.authV1.Get(name)
}

func (r *registry) RegisterUserFactoryV1(name config.DomainConfig_Driver, factory UserFactoryV1) {
	r.userV1.Set(name, factory)
}

func (r *registry) GetUserFactoryV1(name config.DomainConfig_Driver) (UserFactoryV1, bool) {
	return r.userV1.Get(name)
}

// RegisterAuthV1Factory registers a new auth factory.
func RegisterAuthV1Factory(name config.DomainConfig_Driver, factory AuthFactoryV1) {
	globalRegistry.RegisterAuthV1Factory(name, factory)
}

// GetAuthV1Factory gets an auth factory.
// If the auth factory is not found, it will return false.
// If the auth factory is found, it will return true and the auth factory.
func GetAuthV1Factory(name config.DomainConfig_Driver) (AuthFactoryV1, bool) {
	return globalRegistry.GetAuthV1Factory(name)
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
