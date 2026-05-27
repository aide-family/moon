package impl

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
)

func NewCaptchaRepository(c *conf.Bootstrap, d *data.Data) (repository.Captcha, error) {
	repoImpl, close, err := newGoddessCaptcha(c.GetGoddessDomain())
	if err != nil {
		return nil, err
	}
	d.AppendClose("captchaRepo", close)
	return &captchaRepository{CaptchaServer: repoImpl}, nil
}

type captchaRepository struct {
	goddessv1.CaptchaServer
}
