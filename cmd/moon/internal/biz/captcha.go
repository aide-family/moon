package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
	"github.com/aide-cloud/moon/pkg/utils/captcha"
)

type CaptchaBiz struct {
	captchaRepo repo.CaptchaRepo
}

func NewCaptchaBiz(captchaRepo repo.CaptchaRepo) *CaptchaBiz {
	return &CaptchaBiz{
		captchaRepo: captchaRepo,
	}
}

// GenerateCaptcha 生成验证码
func (l *CaptchaBiz) GenerateCaptcha(ctx context.Context, params *bo.GenerateCaptchaParams) (*bo.CaptchaItem, error) {
	id, base64s, err := captcha.CreateCode(ctx, params.Type, params.Theme, params.Size...)
	if err != nil {
		log.Warnw("fun", "captcha.CreateCode", "err", err)
		return nil, merr.ErrorAlert("获取验证码失败")
	}
	// 过期时间
	duration := time.Minute * 1
	expireAt := time.Now().Add(duration).Unix()
	validateCaptchaItem := bo.ValidateCaptchaItem{
		ValidateCaptchaParams: bo.ValidateCaptchaParams{
			Id:    id,
			Value: captcha.GetCodeAnswer(id),
		},
		ExpireAt: expireAt,
	}
	// 存储验证码信息到缓存
	if err = l.captchaRepo.CreateCaptcha(ctx, &validateCaptchaItem, duration); err != nil {
		log.Warnw("fun", "captchaRepo.CreateCaptcha", "err", err)
		return nil, merr.ErrorAlert("获取验证码失败")
	}
	return &bo.CaptchaItem{
		ValidateCaptchaItem: validateCaptchaItem,
		Base64s:             base64s,
	}, nil
}

// VerifyCaptcha 验证验证码
func (l *CaptchaBiz) VerifyCaptcha(ctx context.Context, params *bo.ValidateCaptchaParams) error {
	// 获取验证码信息
	validateCaptchaItem, err := l.captchaRepo.GetCaptchaById(ctx, params.Id)
	if err != nil {
		log.Warnw("fun", "captchaRepo.GetCaptchaById", "err", err)
		return merr.ErrorAlert("验证码已失效")
	}
	// 验证码是否过期
	if time.Now().Unix() > validateCaptchaItem.ExpireAt {
		return merr.ErrorAlert("验证码已失效")
	}
	if validateCaptchaItem.Value != params.Value {
		return merr.ErrorAlert("验证码错误")
	}
	return nil
}
