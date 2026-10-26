package repository

import (
	"context"
	"time"
)

// EmailCodeRepo 保存邮箱验证码的摘要，并保证一次校验后失效。
// 注册码、登录码和绑定码共用这套存储，靠 purpose 隔离；节流按邮箱统计最近一次发码时间。
type EmailCodeRepo interface {
	SaveEmailCode(ctx context.Context, email, codeHash, purpose string, expiresAt time.Time) error
	// SaveEmailCodeWithBinding 同 SaveEmailCode，额外记录第三方绑定凭证摘要，
	// 用于限制同一凭证可发往的不同邮箱数量。
	SaveEmailCodeWithBinding(ctx context.Context, email, codeHash, purpose, bindTokenHash string, expiresAt time.Time) error
	// CountBindSendEmails 统计同一第三方绑定凭证在窗口内已发过码的不同邮箱数量。
	CountBindSendEmails(ctx context.Context, bindTokenHash string, since time.Time) (int, error)
	ConsumeEmailCode(ctx context.Context, email, codeHash, purpose string, now time.Time) (bool, error)
	// IncrementEmailCodeAttempts 记录一次错误校验；达到 maxAttempts 时立即作废该邮箱当前用途的码。
	IncrementEmailCodeAttempts(ctx context.Context, email, purpose string, now time.Time, maxAttempts int) error
	// RecentExhaustedEmailCodes 统计 since 之后因错误次数用尽而作废的码数量，用于邮箱级锁定。
	RecentExhaustedEmailCodes(ctx context.Context, email string, since time.Time, maxAttempts int) (int, error)
	// LastEmailCodeIssuedAt 返回该邮箱最近一次发码时间。
	// 已使用和已过期的码也算发过：节流限制的是发码频率，不是有效码数量。
	// 没有发过码时 ok 为 false。
	LastEmailCodeIssuedAt(ctx context.Context, email string) (issuedAt time.Time, ok bool, err error)
}
