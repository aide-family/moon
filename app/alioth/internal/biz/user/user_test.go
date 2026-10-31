package user

import (
	"context"
	"log/slog"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func TestValidateUsername(t *testing.T) {
	valid := []string{"abc", "新用户", "user1234", "12345678", "名字abc"}
	for _, name := range valid {
		if err := ValidateUsername(name); err != nil {
			t.Fatalf("合法用户名 %q 被拒绝: %v", name, err)
		}
	}
	invalid := []string{"", "abcdefghi", "bad-name", "bad name", "名字!", "emoji😀"}
	for _, name := range invalid {
		if err := ValidateUsername(name); err == nil {
			t.Fatalf("非法用户名 %q 通过了校验", name)
		}
	}
}

func TestDefaultUsername(t *testing.T) {
	cases := map[string]string{
		"new@example.com":            "new",
		"john.doe123456@example.com": "johndoe1",
		"verylongprefix@example.com": "verylong",
		"___@example.com":            "user",
		"用户.名字@example.com":          "用户名字",
		"plain":                      "plain",
	}
	for email, want := range cases {
		got := DefaultUsername(email)
		if got != want {
			t.Fatalf("DefaultUsername(%q) = %q, want %q", email, got, want)
		}
		if err := ValidateUsername(got); err != nil {
			t.Fatalf("DefaultUsername(%q) = %q 不合规: %v", email, got, err)
		}
	}
}

// deleteUserRepoStub 用内存账号表代替数据库，记录 DeleteUser 的调用参数。
type deleteUserRepoStub struct {
	repository.UserRepo
	accounts  map[string]bo.Account
	deleted   []string
	deleteErr error
}

func (stub *deleteUserRepoStub) UserByUID(_ context.Context, uid string) (*bo.Account, error) {
	account, ok := stub.accounts[uid]
	if !ok {
		return nil, merr.NotFound("not found")
	}
	copied := account
	return &copied, nil
}

func (stub *deleteUserRepoStub) DeleteUser(_ context.Context, uid string) error {
	if stub.deleteErr != nil {
		return stub.deleteErr
	}
	stub.deleted = append(stub.deleted, uid)
	return nil
}

// newDeleteDirectory 造一个带平台权限检查的 Directory。
func newDeleteDirectory(accounts map[string]bo.Account) (*Directory, *deleteUserRepoStub) {
	users := &deleteUserRepoStub{accounts: accounts}
	checker := access.NewChecker(users, nil, nil, nil, slog.Default())
	return NewDirectory(users, checker), users
}

func assertDeleteClientMessage(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want error %q, got nil", message)
	}
	if kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("err = %v, want client error", err)
	}
	if got := kerrors.FromError(err).Message; got != message {
		t.Fatalf("message = %q, want %q", got, message)
	}
}

// TestDeleteUserGuards 覆盖删除用户的平台校验：不能删自己、不能删超级管理员、
// 需要 system:user:write；通过后调用仓储删除。
func TestDeleteUserGuards(t *testing.T) {
	actorUID := "actor-1"
	accounts := func() map[string]bo.Account {
		return map[string]bo.Account{
			"actor-1":  {UID: "actor-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSystemAdmin},
			"super-1":  {UID: "super-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSuperAdmin},
			"normal-1": {UID: "normal-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser},
			"reader-1": {UID: "reader-1", Status: "ACTIVE", SystemRole: perm.SystemRoleReadOnly},
		}
	}

	t.Run("rejects deleting yourself", func(t *testing.T) {
		directory, users := newDeleteDirectory(accounts())
		ctx := actor.WithUser(context.Background(), actorUID)
		err := directory.DeleteUser(ctx, actorUID)
		assertDeleteClientMessage(t, err, "cannot delete yourself")
		if len(users.deleted) != 0 {
			t.Fatalf("拒绝删除时不应调用仓储: %v", users.deleted)
		}
	})

	t.Run("rejects deleting super admin", func(t *testing.T) {
		directory, users := newDeleteDirectory(accounts())
		ctx := actor.WithUser(context.Background(), actorUID)
		err := directory.DeleteUser(ctx, "super-1")
		if kerrors.Reason(err) != merr.ReasonForbidden {
			t.Fatalf("err = %v, want forbidden", err)
		}
		if got := kerrors.FromError(err).Message; got != "cannot delete super admin" {
			t.Fatalf("message = %q", got)
		}
		if len(users.deleted) != 0 {
			t.Fatalf("拒绝删除时不应调用仓储: %v", users.deleted)
		}
	})

	t.Run("requires system user write", func(t *testing.T) {
		directory, users := newDeleteDirectory(accounts())
		ctx := actor.WithUser(context.Background(), "reader-1")
		err := directory.DeleteUser(ctx, "normal-1")
		if kerrors.Reason(err) != merr.ReasonForbidden {
			t.Fatalf("err = %v, want forbidden", err)
		}
		if len(users.deleted) != 0 {
			t.Fatalf("没有写权限时不应调用仓储: %v", users.deleted)
		}
	})

	t.Run("deletes ordinary user", func(t *testing.T) {
		directory, users := newDeleteDirectory(accounts())
		ctx := actor.WithUser(context.Background(), actorUID)
		if err := directory.DeleteUser(ctx, "normal-1"); err != nil {
			t.Fatalf("DeleteUser: %v", err)
		}
		if len(users.deleted) != 1 || users.deleted[0] != "normal-1" {
			t.Fatalf("DeleteUser 调用 = %v, want [normal-1]", users.deleted)
		}
	})

	t.Run("wraps storage rejection", func(t *testing.T) {
		directory, users := newDeleteDirectory(accounts())
		users.deleteErr = merr.Client("cannot delete a namespace owner")
		ctx := actor.WithUser(context.Background(), actorUID)
		err := directory.DeleteUser(ctx, "normal-1")
		assertDeleteClientMessage(t, err, "cannot delete a namespace owner")
	})
}
