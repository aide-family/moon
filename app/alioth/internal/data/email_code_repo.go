package data

import (
	"context"
	"crypto/subtle"
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

func (r *emailCodeRepo) ConsumeEmailCode(ctx context.Context, email, codeHash string, now time.Time) (bool, error) {
	rows, err := r.client.EmailLoginCode.Query().Where(
		emaillogincode.EmailEQ(email),
		emaillogincode.UsedAtIsNil(),
		emaillogincode.ExpiresAtGT(now),
	).All(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	for _, row := range rows {
		if subtle.ConstantTimeCompare([]byte(row.CodeHash), []byte(codeHash)) != 1 {
			continue
		}
		if err := r.client.EmailLoginCode.UpdateOne(row).SetUsedAt(now).Exec(ctx); err != nil {
			return false, r.fail(err)
		}
		return true, nil
	}
	return false, nil
}
