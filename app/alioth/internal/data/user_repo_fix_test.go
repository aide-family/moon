package data

import (
	"context"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/pkg/merr"
)

// fixUserRepoStore 打开独立测试库并建表，未设置 ALIOTH_TEST_DSN 时整组测试跳过。
func fixUserRepoStore(t *testing.T) *Dependencies {
	t.Helper()
	source := testDSN(t)
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := ApplySchema(context.Background(), store.client, store.db); err != nil {
		t.Fatalf("ApplySchema: %v", err)
	}
	return store
}

// createFixUser 直接造一条普通用户，并在测试结束后按本测试的 UID 清理。
func createFixUser(t *testing.T, store *Dependencies, uid string) {
	t.Helper()
	row, err := store.client.User.Create().
		SetUID(uid).
		SetName("fix-user").
		SetNickname("fix-user").
		SetEmail(uid + "@test.example.com").
		SetSystemRole(user.SystemRoleUSER).
		Save(context.Background())
	if err != nil {
		t.Fatalf("创建用户 %s: %v", uid, err)
	}
	t.Cleanup(func() {
		_, _ = store.client.User.Delete().Where(user.UIDEQ(row.UID)).Exec(context.Background())
	})
}

// TestUpdateUserSystemRoleReadOnly 复现 M24/M1：平台只读角色在 biz 与 proto 都允许，
// 数据层也必须接受，并且写入授权人，否则接口返回 400 invalid request。
func TestUpdateUserSystemRoleReadOnly(t *testing.T) {
	store := fixUserRepoStore(t)
	ctx := context.Background()
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	uid := "fix-user-" + suffix
	operatorUID := "fix-operator-" + suffix
	createFixUser(t, store, uid)

	users := NewUserRepo(store, slog.Default())
	account, err := users.UpdateUserSystemRole(ctx, uid, "SYSTEM_READONLY", operatorUID)
	if err != nil {
		t.Fatalf("UpdateUserSystemRole(SYSTEM_READONLY) error = %v", err)
	}
	if account.SystemRole != "SYSTEM_READONLY" || account.AuthorizedBy != operatorUID {
		t.Fatalf("account = %+v", account)
	}

	reloaded, err := users.UserByUID(ctx, uid)
	if err != nil {
		t.Fatalf("UserByUID error = %v", err)
	}
	if reloaded.SystemRole != "SYSTEM_READONLY" || reloaded.AuthorizedBy != operatorUID {
		t.Fatalf("reloaded = %+v", reloaded)
	}
}

// TestUpdateUserSystemRoleKeepsRejectingUnknown 校验非法平台角色仍被拒绝，且不落库。
func TestUpdateUserSystemRoleKeepsRejectingUnknown(t *testing.T) {
	store := fixUserRepoStore(t)
	ctx := context.Background()
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	uid := "fix-user-invalid-" + suffix
	operatorUID := "fix-operator-" + suffix
	createFixUser(t, store, uid)

	users := NewUserRepo(store, slog.Default())
	for _, role := range []string{"SUPER_ADMIN", "ROOT", ""} {
		if _, err := users.UpdateUserSystemRole(ctx, uid, role, operatorUID); !merr.IsInvalid(err) {
			t.Fatalf("UpdateUserSystemRole(%q) error = %v, want invalid", role, err)
		}
	}

	reloaded, err := users.UserByUID(ctx, uid)
	if err != nil {
		t.Fatalf("UserByUID error = %v", err)
	}
	if reloaded.SystemRole != string(user.SystemRoleUSER) || reloaded.AuthorizedBy != "" {
		t.Fatalf("reloaded after invalid roles = %+v", reloaded)
	}
	if _, err := users.UpdateUserSystemRole(ctx, uid, "SYSTEM_ADMIN", operatorUID); err != nil {
		t.Fatalf("UpdateUserSystemRole(SYSTEM_ADMIN) error = %v", err)
	}
	reloaded, err = users.UserByUID(ctx, uid)
	if err != nil {
		t.Fatalf("UserByUID error = %v", err)
	}
	if reloaded.SystemRole != "SYSTEM_ADMIN" || reloaded.AuthorizedBy != operatorUID {
		t.Fatalf("reloaded after valid role = %+v", reloaded)
	}
}
