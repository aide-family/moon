// Package captcha provides domain factory registration for the captcha service.
package captcha

import (
	"github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/config"
	domainregister "github.com/aide-family/magicbox/domain"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new captcha registry.
func NewRegistry() Registry {
	return &registry{
		captchaV1: domainregister.NewRegistry[CaptchaFactoryV1](),
	}
}

// CaptchaFactoryV1 is the factory function for the captcha service.
type CaptchaFactoryV1 func(c *config.DomainConfig) (v1.CaptchaServer, func() error, error)

type Registry interface {
	RegisterCaptchaFactoryV1(name config.DomainConfig_Driver, factory CaptchaFactoryV1)
	GetCaptchaFactoryV1(name config.DomainConfig_Driver) (CaptchaFactoryV1, bool)
}

type registry struct {
	captchaV1 *domainregister.Registry[CaptchaFactoryV1]
}

func (r *registry) RegisterCaptchaFactoryV1(name config.DomainConfig_Driver, factory CaptchaFactoryV1) {
	r.captchaV1.Register(name, factory)
}

func (r *registry) GetCaptchaFactoryV1(name config.DomainConfig_Driver) (CaptchaFactoryV1, bool) {
	return r.captchaV1.Get(name)
}

// RegisterCaptchaFactoryV1 registers a new captcha factory.
func RegisterCaptchaFactoryV1(name config.DomainConfig_Driver, factory CaptchaFactoryV1) {
	globalRegistry.RegisterCaptchaFactoryV1(name, factory)
}

// GetCaptchaFactoryV1 gets a captcha factory.
func GetCaptchaFactoryV1(name config.DomainConfig_Driver) (CaptchaFactoryV1, bool) {
	return globalRegistry.GetCaptchaFactoryV1(name)
}

