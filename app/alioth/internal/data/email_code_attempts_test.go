package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/emaillogincode"
)

// TestEmailCodeAttemptsExhaustAndLockout 验证错误次数达到上限后码立即作废，
// 并能在窗口内统计被作废的码数量（用于邮箱级锁定）。
func TestEmailCodeAttemptsExhaustAndLockout(t *testing.T) {
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
	email := "attempts-" + time.Now().Format("150405.000000000") + "@example.test"
	t.Cleanup(func() {
		_, _ = store.client.EmailLoginCode.Delete().Where(emaillogincode.EmailEQ(email)).Exec(ctx)
	})
	now := time.Now()
	const maxAttempts = 5

	// 五次错误后码被作废，正确答案也无法再消费。
	if err := repo.SaveEmailCode(ctx, email, "login-hash", "login", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxAttempts; i++ {
		if err := repo.IncrementEmailCodeAttempts(ctx, email, "login", now, maxAttempts); err != nil {
			t.Fatal(err)
		}
	}
	if ok, err := repo.ConsumeEmailCode(ctx, email, "login-hash", "login", now); err != nil || ok {
		t.Fatalf("exhausted code must not be consumable: ok=%v err=%v", ok, err)
	}
	count, err := repo.RecentExhaustedEmailCodes(ctx, email, now.Add(-time.Minute), maxAttempts)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("exhausted count = %d, want 1", count)
	}

	// 新码重新计数；再作废一次后窗口内计数为 2。
	if err := repo.SaveEmailCode(ctx, email, "login-hash-2", "login", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxAttempts; i++ {
		if err := repo.IncrementEmailCodeAttempts(ctx, email, "login", now, maxAttempts); err != nil {
			t.Fatal(err)
		}
	}
	count, err = repo.RecentExhaustedEmailCodes(ctx, email, now.Add(-time.Minute), maxAttempts)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("exhausted count = %d, want 2", count)
	}

	// 窗口外的作废码不计入，锁定会随时间自动解除。
	count, err = repo.RecentExhaustedEmailCodes(ctx, email, now.Add(time.Minute), maxAttempts)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("exhausted count outside window = %d, want 0", count)
	}
}
