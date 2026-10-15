package data

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
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

func (r *captchaRepo) ConsumeCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error) {
	row, err := r.client.Captcha.Query().Where(captcha.UIDEQ(id), captcha.UsedAtIsNil()).Only(ctx)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, r.fail(err)
	}
	if now.After(row.ExpiresAt) || subtle.ConstantTimeCompare([]byte(row.AnswerHash), []byte(answerHash)) != 1 {
		return false, nil
	}
	if err := r.client.Captcha.UpdateOne(row).SetUsedAt(now).Exec(ctx); err != nil {
		return false, r.fail(err)
	}
	return true, nil
}
