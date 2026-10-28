package data

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/emaillogincode"
)

// TestBindSendAddressCap 同一绑定凭证只统计不同邮箱；重复地址不增加计数，其他凭证互不影响。
func TestBindSendAddressCap(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	repo := NewEmailCodeRepo(store, slog.Default())
	suffix := time.Now().Format("150405.000000000")
	tokenHash := "token-" + suffix
	otherHash := "other-" + suffix
	emails := make([]string, 0, 5)
	t.Cleanup(func() {
		_, _ = store.client.EmailLoginCode.Delete().Where(emaillogincode.BindTokenHashIn(tokenHash, otherHash)).Exec(ctx)
	})
	now := time.Now()

	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("bind-%s-%d@example.test", suffix, i)
		emails = append(emails, email)
		if err := repo.SaveEmailCodeWithBinding(ctx, email, "hash", "bind", tokenHash, now.Add(10*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	// 重复地址不增加计数。
	if err := repo.SaveEmailCodeWithBinding(ctx, emails[0], "hash-2", "bind", tokenHash, now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	count, err := repo.CountBindSendEmails(ctx, tokenHash, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Fatalf("distinct bind emails = %d, want 5", count)
	}
	if count, err := repo.CountBindSendEmails(ctx, otherHash, now.Add(-time.Hour)); err != nil || count != 0 {
		t.Fatalf("other token count = %d err=%v, want 0", count, err)
	}
}
