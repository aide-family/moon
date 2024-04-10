package captcha

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/helper/consts"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
)

var _ repository.CaptchaRepo = (*captchaRepoImpl)(nil)

type captchaRepoImpl struct {
	repository.UnimplementedCaptchaRepo

	data *data.Data
	log  *log.Helper
}

func (l *captchaRepoImpl) CreateCaptcha(ctx context.Context, captcha *bo.CaptchaBO) error {
	key := consts.AuthCaptchaKey.Key(captcha.Id).String()
	return l.data.Cache().Set(ctx, key, captcha.Bytes(), time.Minute*3)
}

func (l *captchaRepoImpl) GetCaptchaById(ctx context.Context, id string) (*bo.CaptchaBO, error) {
	key := consts.AuthCaptchaKey.Key(id).String()
	var captcha bo.CaptchaBO
	value, err := l.data.Cache().Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(value, &captcha); err != nil {
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
