package captchav1

import (
	"github.com/aide-family/magicbox/config"
	"google.golang.org/protobuf/types/known/anypb"

	captchadomain "github.com/aide-family/goddess/domain/captcha"
	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

func init() {
	captchadomain.RegisterCaptchaFactoryV1(config.DomainConfig_DEFAULT, NewDefaultCaptcha)
}

// NewDefaultCaptcha creates the default captcha server (DEFAULT driver, no DB required).
func NewDefaultCaptcha(_ *config.DomainConfig, driver *anypb.Any) (goddessv1.CaptchaServer, func() error, error) {
	captchaBiz := biz.NewCaptcha()
	return &defaultCaptcha{CaptchaServer: service.NewCaptchaService(captchaBiz)}, func() error { return nil }, nil
}

type defaultCaptcha struct {
	goddessv1.CaptchaServer
}
