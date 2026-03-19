// Package captcha provides domain factory registration for the captcha service.
package captcha

import (
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"

	v1 "github.com/aide-family/goddess/pkg/api/v1"
)

var globalRegistry = newRegistry()

// newRegistry creates a new captcha registry.
func newRegistry() *registry {
	return &registry{
		captchaV1: domainregister.NewRegistry[CaptchaFactoryV1](),
	}
}

// CaptchaFactoryV1 is the factory function for the captcha service.
type CaptchaFactoryV1 func(c *config.DomainConfig) (v1.CaptchaServer, func() error, error)

type registry struct {
	captchaV1 *domainregister.Registry[CaptchaFactoryV1]
}

// RegisterCaptchaFactoryV1 registers a new captcha factory.
func RegisterCaptchaFactoryV1(name config.DomainConfig_Driver, factory CaptchaFactoryV1) {
	globalRegistry.captchaV1.Register(name, factory)
}

// GetCaptchaFactoryV1 gets a captcha factory.
// If the captcha factory is not found, it will return false.
// If the captcha factory is found, it will return true and the captcha factory.
func GetCaptchaFactoryV1(name config.DomainConfig_Driver) (CaptchaFactoryV1, bool) {
	return globalRegistry.captchaV1.Get(name)
}
