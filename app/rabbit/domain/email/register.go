// Package email provides domain factory registration for the email service.
package email

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
	"google.golang.org/protobuf/types/known/anypb"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new email registry.
func newRegistry() *registry {
	return &registry{
		emailV1: domainregister.NewRegistry[EmailFactoryV1](),
	}
}

// EmailFactoryV1 is the factory function for the email service.
type EmailFactoryV1 func(c *config.DomainConfig, driver *anypb.Any) (apiv1.EmailServer, func() error, error)

type registry struct {
	emailV1 *domainregister.Registry[EmailFactoryV1]
}

// RegisterEmailV1Factory registers a new email factory.
func RegisterEmailV1Factory(name config.DomainConfig_Driver, factory EmailFactoryV1) {
	globalRegistry.emailV1.Register(name, factory)
}

// GetEmailV1Factory gets a email factory.
// If the email factory is not found, it will return false.
// If the email factory is found, it will return true and the email factory.
func GetEmailV1Factory(name config.DomainConfig_Driver) (EmailFactoryV1, bool) {
	return globalRegistry.emailV1.Get(name)
}

