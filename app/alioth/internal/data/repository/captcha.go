package repository

import (
	"context"
	"time"
)

// CaptchaRepo 保存图形验证码的答案摘要，并保证一次校验后失效。
type CaptchaRepo interface {
	SaveCaptcha(ctx context.Context, id, answerHash string, expiresAt time.Time) error
	ConsumeCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error)
}
