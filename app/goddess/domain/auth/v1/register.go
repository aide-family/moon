// Package authv1 is the auth service implementation.
package authv1

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new auth registry.
func NewRegistry() Registry {
	return &registry{
		authV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]AuthFactoryV1)),
	}
}

type (
	Registry interface {
		RegisterAuthV1Factory(name config.DomainConfig_Driver, factory AuthFactoryV1)
		GetAuthV1Factory(name config.DomainConfig_Driver) (AuthFactoryV1, bool)
	}

	// AuthFactoryV1 is the factory function for the auth service.
	AuthFactoryV1 func(c *config.DomainConfig, jwtConfig *config.JWT) (goddessv1.AuthServiceServer, func() error, error)
)

type registry struct {
	authV1 *safety.SyncMap[config.DomainConfig_Driver, AuthFactoryV1]
}

func (r *registry) RegisterAuthV1Factory(name config.DomainConfig_Driver, factory AuthFactoryV1) {
	r.authV1.Set(name, factory)
}

func (r *registry) GetAuthV1Factory(name config.DomainConfig_Driver) (AuthFactoryV1, bool) {
	return r.authV1.Get(name)
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
