package repository

import (
	"context"
	"time"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
)

type Captcha interface {
	// CreateCaptcha 创建验证码
	CreateCaptcha(ctx context.Context, captcha *bo.ValidateCaptchaItem, duration time.Duration) error
	// GetCaptchaById 通过id获取验证码详情
	GetCaptchaById(ctx context.Context, id string) (*bo.ValidateCaptchaItem, error)
}
