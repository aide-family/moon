package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/emaillogincode"
)

// TestEmailCodePurposeIsolation 验证码按用途隔离：绑定码不能用于登录，登录码不能用于绑定；
// 升级前没有 purpose 的历史行按通配保留，避免部署后 10 分钟内的有效码全部失效。
func TestEmailCodePurposeIsolation(t *testing.T) {
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
	email := "purpose-" + time.Now().Format("150405.000000000") + "@example.test"
	t.Cleanup(func() {
		_, _ = store.client.EmailLoginCode.Delete().Where(emaillogincode.EmailEQ(email)).Exec(ctx)
	})
	now := time.Now()

	if err := repo.SaveEmailCode(ctx, email, "bind-hash", "bind", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.ConsumeEmailCode(ctx, email, "bind-hash", "login", now); err != nil || ok {
		t.Fatalf("bind code must not be consumable as login: ok=%v err=%v", ok, err)
	}
	if ok, err := repo.ConsumeEmailCode(ctx, email, "bind-hash", "bind", now); err != nil || !ok {
		t.Fatalf("bind code must be consumable as bind: ok=%v err=%v", ok, err)
	}

	// 历史空 purpose 行按通配匹配。
	if err := repo.SaveEmailCode(ctx, email, "legacy-hash", "", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.ConsumeEmailCode(ctx, email, "legacy-hash", "login", now); err != nil || !ok {
		t.Fatalf("legacy empty-purpose code should still match: ok=%v err=%v", ok, err)
	}

	// 新码只作废同用途旧码：换绑码不应顶掉还没用的登录码，反之亦然。
	if err := repo.SaveEmailCode(ctx, email, "login-hash", "login", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveEmailCode(ctx, email, "bind2-hash", "bind", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.ConsumeEmailCode(ctx, email, "login-hash", "login", now); err != nil || !ok {
		t.Fatalf("login code should survive a later bind code save: ok=%v err=%v", ok, err)
	}
	if ok, err := repo.ConsumeEmailCode(ctx, email, "bind2-hash", "bind", now); err != nil || !ok {
		t.Fatalf("bind code should survive a later login code save: ok=%v err=%v", ok, err)
	}
}
