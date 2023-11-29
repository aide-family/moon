package captcha

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper"
)

var _ repository.CaptchaRepo = (*captchaRepoImpl)(nil)

type captchaRepoImpl struct {
	repository.UnimplementedCaptchaRepo

	data *data.Data
	log  *log.Helper
}

func (l *captchaRepoImpl) CreateCaptcha(ctx context.Context, captcha *dobo.CaptchaDO) error {
	key := helper.AuthCaptchaKey.Key(captcha.Id).String()
	return l.data.Client().Set(ctx, key, captcha, time.Minute*3).Err()
}

func (l *captchaRepoImpl) GetCaptchaById(ctx context.Context, id string) (*dobo.CaptchaDO, error) {
	key := helper.AuthCaptchaKey.Key(id).String()
	var captcha dobo.CaptchaDO
	if err := l.data.Client().Get(ctx, key).Scan(&captcha); err != nil {
		return nil, err
	}

	return &captcha, nil
}

func NewCaptchaRepo(data *data.Data, logger log.Logger) repository.CaptchaRepo {
	return &captchaRepoImpl{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "repository.captcha")),
	}
}
