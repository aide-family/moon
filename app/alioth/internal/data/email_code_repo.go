package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/emaillogincode"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

func (r *emailCodeRepo) SaveEmailCode(ctx context.Context, email, codeHash, purpose string, expiresAt time.Time) error {
	return r.saveEmailCode(ctx, email, codeHash, purpose, "", expiresAt)
}

func (r *emailCodeRepo) SaveEmailCodeWithBinding(ctx context.Context, email, codeHash, purpose, bindTokenHash string, expiresAt time.Time) error {
	return r.saveEmailCode(ctx, email, codeHash, purpose, bindTokenHash, expiresAt)
}

func (r *emailCodeRepo) saveEmailCode(ctx context.Context, email, codeHash, purpose, bindTokenHash string, expiresAt time.Time) error {
	now := time.Now()
	if _, err := r.client.EmailLoginCode.Update().
		Where(
			emaillogincode.EmailEQ(email),
			emaillogincode.UsedAtIsNil(),
			// 只作废同用途的旧码：换绑邮箱的码不应该顶掉还没用的登录码。
			emaillogincode.Or(
				emaillogincode.PurposeEQ(purpose),
				emaillogincode.PurposeEQ(""),
			),
		).
		SetUsedAt(now).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	_, err := r.client.EmailLoginCode.Create().
		SetUID(id.New()).
		SetEmail(email).
		SetCodeHash(codeHash).
		SetPurpose(purpose).
		SetBindTokenHash(bindTokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return r.fail(err)
}

// CountBindSendEmails 统计同一绑定凭证在窗口内已发过码的不同邮箱数量。
// 同一邮箱重复请求只算一个地址。
func (r *emailCodeRepo) CountBindSendEmails(ctx context.Context, bindTokenHash string, since time.Time) (int, error) {
	if bindTokenHash == "" {
		return 0, nil
	}
	rows, err := r.client.EmailLoginCode.Query().
		Where(
			emaillogincode.BindTokenHashEQ(bindTokenHash),
			emaillogincode.CreatedAtGTE(since),
		).
		Select(emaillogincode.FieldEmail).
		All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		seen[row.Email] = struct{}{}
	}
	return len(seen), nil
}

// ConsumeEmailCode 原子消费邮箱验证码。
// 与图形验证码同理：条件全部放进 UPDATE，避免并发下同一个码被用两次。
// 空 purpose 的行是升级前的历史数据，按通配处理，避免部署时把 10 分钟内有效的码全部作废。
func (r *emailCodeRepo) ConsumeEmailCode(ctx context.Context, email, codeHash, purpose string, now time.Time) (bool, error) {
	affected, err := r.client.EmailLoginCode.Update().
		Where(
			emaillogincode.EmailEQ(email),
			emaillogincode.UsedAtIsNil(),
			emaillogincode.ExpiresAtGT(now),
			emaillogincode.CodeHashEQ(codeHash),
			emaillogincode.Or(
				emaillogincode.PurposeEQ(purpose),
				emaillogincode.PurposeEQ(""),
			),
		).
		SetUsedAt(now).
		Save(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return affected >= 1, nil
}

// IncrementEmailCodeAttempts 记录一次错误校验。达到上限的码立即作废，
// 避免同一个 6 位码在 10 分钟有效期内被无限次尝试。
func (r *emailCodeRepo) IncrementEmailCodeAttempts(ctx context.Context, email, purpose string, now time.Time, maxAttempts int) error {
	if _, err := r.client.EmailLoginCode.Update().
		Where(
			emaillogincode.EmailEQ(email),
			emaillogincode.UsedAtIsNil(),
			emaillogincode.ExpiresAtGT(now),
			emaillogincode.Or(
				emaillogincode.PurposeEQ(purpose),
				emaillogincode.PurposeEQ(""),
			),
		).
		AddAttempts(1).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if maxAttempts <= 0 {
		return nil
	}
	if _, err := r.client.EmailLoginCode.Update().
		Where(
			emaillogincode.EmailEQ(email),
			emaillogincode.UsedAtIsNil(),
			emaillogincode.ExpiresAtGT(now),
			emaillogincode.AttemptsGTE(maxAttempts),
		).
		SetUsedAt(now).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

// RecentExhaustedEmailCodes 统计窗口内因错误次数用尽而作废的码数量。
// 正常消费成功的码 attempts 不会达到阈值，因此不会被计入。
func (r *emailCodeRepo) RecentExhaustedEmailCodes(ctx context.Context, email string, since time.Time, maxAttempts int) (int, error) {
	count, err := r.client.EmailLoginCode.Query().
		Where(
			emaillogincode.EmailEQ(email),
			emaillogincode.AttemptsGTE(maxAttempts),
			emaillogincode.UsedAtNotNil(),
			emaillogincode.UsedAtGTE(since),
		).
		Count(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return count, nil
}

// LastEmailCodeIssuedAt 返回该邮箱最近一次发码时间，用于发码节流。
// 用 created_at 判断：已使用和已过期的码也算发过，节流描述的是发码频率。
func (r *emailCodeRepo) LastEmailCodeIssuedAt(ctx context.Context, email string) (time.Time, bool, error) {
	record, err := r.client.EmailLoginCode.Query().
		Where(emaillogincode.EmailEQ(email)).
		Order(ent.Desc(emaillogincode.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, r.fail(err)
	}
	return record.CreatedAt, true, nil
}
