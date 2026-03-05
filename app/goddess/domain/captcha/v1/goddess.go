package captchav1

import (
	"github.com/aide-family/magicbox/config"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	RegisterCaptchaFactoryV1(config.DomainConfig_DEFAULT, NewDefaultCaptcha)
}

// NewDefaultCaptcha creates the default captcha server (no DB required).
func NewDefaultCaptcha(_ *config.DomainConfig) (goddessv1.CaptchaServer, func() error, error) {
	captchaBiz := biz.NewCaptcha()
	return &captchaRepository{CaptchaServer: service.NewCaptchaService(captchaBiz)}, func() error { return nil }, nil
}

type captchaRepository struct {
	goddessv1.CaptchaServer
}
