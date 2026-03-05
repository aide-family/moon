// Package captchav1 is the captcha service implementation.
package captchav1

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new captcha registry.
func NewRegistry() Registry {
	return &registry{
		captchaV1: safety.NewSyncMap(make(map[config.DomainConfig_Driver]CaptchaFactoryV1)),
	}
}

// CaptchaFactoryV1 is the factory function for the captcha service.
type (
	CaptchaFactoryV1 func(c *config.DomainConfig) (goddessv1.CaptchaServer, func() error, error)
	Registry         interface {
		RegisterCaptchaFactoryV1(name config.DomainConfig_Driver, factory CaptchaFactoryV1)
		GetCaptchaFactoryV1(name config.DomainConfig_Driver) (CaptchaFactoryV1, bool)
	}
)

type registry struct {
	captchaV1 *safety.SyncMap[config.DomainConfig_Driver, CaptchaFactoryV1]
}

func (r *registry) RegisterCaptchaFactoryV1(name config.DomainConfig_Driver, factory CaptchaFactoryV1) {
	r.captchaV1.Set(name, factory)
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
