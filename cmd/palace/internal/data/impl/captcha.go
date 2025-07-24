package impl

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mojocn/base64Captcha"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/captcha"
)

func NewCaptchaRepo(bc *conf.Bootstrap, d *data.Data, logger log.Logger) repository.Captcha {
	captchaConf := bc.GetAuth().GetCaptcha()
	captchaStoreOpts := []captcha.StoreOption{
		captcha.WithExpire(captchaConf.GetExpire().AsDuration()),
		captcha.WithTimeout(captchaConf.GetTimeout().AsDuration()),
		captcha.WithPrefix(captchaConf.GetPrefix()),
	}
	captchaStore := captcha.NewStore(d.GetCache().Client(), captchaStoreOpts...)
	captchaInstance := captcha.NewCaptcha(captcha.WithStore(captchaStore))
	return &captchaRepoImpl{
		expired: captchaConf.GetExpire().AsDuration(),
		Data:    d,
		Captcha: captchaInstance,
		helper:  log.NewHelper(log.With(logger, "module", "data.repo.captcha")),
	}
}

type captchaRepoImpl struct {
	*data.Data
	*base64Captcha.Captcha
	expired time.Duration

	helper *log.Helper
}

func (c *captchaRepoImpl) Generate(_ context.Context) (*bo.Captcha, error) {
	id, b64s, answer, err := c.Captcha.Generate()
	if err != nil {
		return nil, err
	}
	return &bo.Captcha{
		ID:             id,
		B64s:           b64s,
		Answer:         answer,
		ExpiredSeconds: int64(c.expired.Seconds()),
	}, nil
}

func (c *captchaRepoImpl) Verify(_ context.Context, req *bo.CaptchaVerify) bool {
	return c.Captcha.Verify(req.ID, req.Answer, req.Clear)
}
