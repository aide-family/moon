package repository

import (
	"context"
	"time"
)

// EmailCodeRepo 保存邮箱验证码的摘要，并保证一次校验后失效。
// 注册码、登录码和绑定码共用这套存储，靠 purpose 隔离；节流按邮箱统计最近一次发码时间。
type EmailCodeRepo interface {
	SaveEmailCode(ctx context.Context, email, codeHash, purpose string, expiresAt time.Time) error
	ConsumeEmailCode(ctx context.Context, email, codeHash, purpose string, now time.Time) (bool, error)
	// LastEmailCodeIssuedAt 返回该邮箱最近一次发码时间。
	// 已使用和已过期的码也算发过：节流限制的是发码频率，不是有效码数量。
	// 没有发过码时 ok 为 false。
	LastEmailCodeIssuedAt(ctx context.Context, email string) (issuedAt time.Time, ok bool, err error)
}
