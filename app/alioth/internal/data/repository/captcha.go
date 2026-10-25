package repository

import (
	"context"
	"time"
)

// CaptchaRepo 保存图形验证码的答案摘要，并保证一次校验后失效。
// 发码前用 CheckCaptcha 校验但不消费，提交成功后再 ConsumeCaptcha：
// 一次解题可以走完发码和提交，而没解题的人既拿不到邮箱码也没法提交。
type CaptchaRepo interface {
	SaveCaptcha(ctx context.Context, id, answerHash string, expiresAt time.Time) error
	// CheckCaptcha 只校验答案是否匹配且未过期、未使用，不改变记录。
	CheckCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error)
	ConsumeCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error)
}
