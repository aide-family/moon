package repository

import (
	"context"
	"time"
)

// EmailCodeRepo 保存邮箱登录码的摘要，并保证一次校验后失效。
type EmailCodeRepo interface {
	SaveEmailCode(ctx context.Context, email, codeHash string, expiresAt time.Time) error
	ConsumeEmailCode(ctx context.Context, email, codeHash string, now time.Time) (bool, error)
}
