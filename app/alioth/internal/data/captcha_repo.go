package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/captcha"
)

func (r *captchaRepo) SaveCaptcha(ctx context.Context, id, answerHash string, expiresAt time.Time) error {
	_, err := r.client.Captcha.Create().
		SetUID(id).
		SetAnswerHash(answerHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return r.fail(err)
}

// ConsumeCaptcha 原子消费一次性图形验证码。
// 先查后改会让并发请求拿到同一个验证码：这里把 used_at IS NULL、过期时间
// 和答案匹配全部下推到 UPDATE 的 WHERE，只有受影响行数为 1 才算消费成功。
func (r *captchaRepo) ConsumeCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error) {
	affected, err := r.client.Captcha.Update().
		Where(
			captcha.UIDEQ(id),
			captcha.UsedAtIsNil(),
			captcha.ExpiresAtGT(now),
			captcha.AnswerHashEQ(answerHash),
		).
		SetUsedAt(now).
		Save(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return affected == 1, nil
}
