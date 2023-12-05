package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/util/captcha"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type (
	CaptchaBiz struct {
		log *log.Helper

		captchaRepo repository.CaptchaRepo
	}
)

func NewCaptchaBiz(captchaRepo repository.CaptchaRepo, logger log.Logger) *CaptchaBiz {
	return &CaptchaBiz{
		log: log.NewHelper(logger),

		captchaRepo: captchaRepo,
	}
}

// GenerateCaptcha 生成验证码
func (b *CaptchaBiz) GenerateCaptcha(ctx context.Context, captchaType captcha.Type, size ...int) (*bo.CaptchaBO, error) {
	codeId, codeImageBase64, err := captcha.CreateCode(ctx, captchaType, size...)
	if err != nil {
		return nil, err
	}
	// 过期时间
	expireAt := time.Now().Add(time.Minute * 1).Unix()
	captchaBo := &bo.CaptchaBO{
		Id:       codeId,
		Value:    captcha.GetCodeAnswer(codeId),
		Image:    codeImageBase64,
		ExpireAt: expireAt,
	}

	// 存储验证码到缓存
	if err = b.captchaRepo.CreateCaptcha(ctx, captchaBo); err != nil {
		return nil, err
	}

	return captchaBo, nil
}

// VerifyCaptcha 验证验证码
func (b *CaptchaBiz) VerifyCaptcha(ctx context.Context, codeId, codeValue string) error {
	captchaDo, err := b.captchaRepo.GetCaptchaById(ctx, codeId)
	if err != nil {
		return perrors.ErrorNotFound("验证码错误")
	}
	// 判断验证码是否过期
	if captchaDo.ExpireAt < time.Now().Unix() {
		return perrors.ErrorInvalidParams("验证码已过期")
	}

	if captchaDo.Value == codeValue {
		return nil
	}

	return perrors.ErrorInvalidParams("验证码错误")
}
