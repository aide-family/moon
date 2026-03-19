package impl

import (
	captchaDomain "github.com/aide-family/goddess/domain/captcha"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewCaptchaRepository(c *conf.Bootstrap, d *data.Data) (repository.Captcha, error) {
	repoConfig := c.GetCaptchaDomain()
	if repoConfig == nil {
		return nil, merr.ErrorInternalServer("captchaDomain is required")
	}
	driver := repoConfig.GetDriver()
	version := repoConfig.GetVersion()
	switch version {
	default:
		factory, ok := captchaDomain.GetCaptchaFactoryV1(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("captcha repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig)
		if err != nil {
			return nil, err
		}
		d.AppendClose("captchaRepo", close)
		return &captchaRepository{CaptchaServer: repoImpl}, nil
	}
}

type captchaRepository struct {
	goddessv1.CaptchaServer
}
