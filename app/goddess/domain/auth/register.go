// Package auth provides domain factory registration for the auth service.
package auth

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	"google.golang.org/protobuf/types/known/anypb"

	v1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new auth registry.
func newRegistry() *registry {
	return &registry{
		authV1: domainregister.NewRegistry[AuthFactoryV1](),
	}
}

// AuthFactoryV1 is the factory function for the auth service.
type AuthFactoryV1 func(c *config.DomainConfig, driver *anypb.Any) (v1.AuthServiceServer, func() error, error)

type registry struct {
	authV1 *domainregister.Registry[AuthFactoryV1]
}

// RegisterAuthV1Factory registers a new auth factory.
func RegisterAuthV1Factory(name config.DomainConfig_Driver, factory AuthFactoryV1) {
	globalRegistry.authV1.Register(name, factory)
}

// GetAuthV1Factory gets an auth factory.
// If the auth factory is not found, it will return false.
// If the auth factory is found, it will return true and the auth factory.
func GetAuthV1Factory(name config.DomainConfig_Driver) (AuthFactoryV1, bool) {
	return globalRegistry.authV1.Get(name)
}
