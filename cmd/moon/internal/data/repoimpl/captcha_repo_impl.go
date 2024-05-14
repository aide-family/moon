package repoimpl

import (
	"context"
	"time"

	"github.com/aide-cloud/moon/cmd/moon/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/moon/internal/data"
)

func NewCaptchaRepo(data *data.Data) repo.CaptchaRepo {
	return &captchaRepoImpl{
		data: data,
	}
}

type captchaRepoImpl struct {
	data *data.Data
}

func (l *captchaRepoImpl) CreateCaptcha(ctx context.Context, captcha *bo.ValidateCaptchaItem, duration time.Duration) error {
	bs, err := captcha.MarshalBinary()
	if err != nil {
		return err
	}
	return l.data.GetCacher().Set(ctx, captcha.Id, string(bs), duration)
}

func (l *captchaRepoImpl) GetCaptchaById(ctx context.Context, id string) (*bo.ValidateCaptchaItem, error) {
	str, err := l.data.GetCacher().Get(ctx, id)
	if err != nil {
		return nil, err
	}
	var captcha bo.ValidateCaptchaItem
	if err := captcha.UnmarshalBinary([]byte(str)); err != nil {
		return nil, err
	}
	return &captcha, nil
}
