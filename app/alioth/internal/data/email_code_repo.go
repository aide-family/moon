package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/emaillogincode"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
)

func (r *emailCodeRepo) SaveEmailCode(ctx context.Context, email, codeHash string, expiresAt time.Time) error {
	now := time.Now()
	if _, err := r.client.EmailLoginCode.Update().
		Where(emaillogincode.EmailEQ(email), emaillogincode.UsedAtIsNil()).
		SetUsedAt(now).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	_, err := r.client.EmailLoginCode.Create().
		SetUID(id.New()).
		SetEmail(email).
		SetCodeHash(codeHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return r.fail(err)
}

// ConsumeEmailCode 原子消费邮箱登录码。
// 与图形验证码同理：条件全部放进 UPDATE，避免并发下同一个码被用两次。
func (r *emailCodeRepo) ConsumeEmailCode(ctx context.Context, email, codeHash string, now time.Time) (bool, error) {
	affected, err := r.client.EmailLoginCode.Update().
		Where(
			emaillogincode.EmailEQ(email),
			emaillogincode.UsedAtIsNil(),
			emaillogincode.ExpiresAtGT(now),
			emaillogincode.CodeHashEQ(codeHash),
		).
		SetUsedAt(now).
		Save(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return affected >= 1, nil
}
