package auth

import (
	"context"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/actor"
)

// TestChangeEmailRequiresVerifiedCode 未带验证码 / 验证码错误时都不写邮箱。
func TestChangeEmailRequiresVerifiedCode(t *testing.T) {
	fixture := newAuthFixture(t)
	uid := "user-change-email"
	fixture.addAccount(uid, "old@example.com")
	ctx := actor.WithUser(context.Background(), uid)

	if _, err := fixture.authenticator.ChangeEmail(ctx, "new@example.com", "", "cap-1", "7"); err == nil {
		t.Fatal("empty code must be rejected")
	}

	fixture.solveCaptcha(t, "cap-1", "7")
	if _, err := fixture.authenticator.ChangeEmail(ctx, "new@example.com", "000000", "cap-1", "7"); err == nil {
		t.Fatal("wrong code must be rejected")
	}
	if got := fixture.users.byUID[uid].Email; got != "old@example.com" {
		t.Fatalf("email changed without verification: %q", got)
	}
}

// TestChangeEmailConsumesNewAddressCode 验证码发到新地址并校验用途；成功后写入新邮箱。
func TestChangeEmailConsumesNewAddressCode(t *testing.T) {
	fixture := newAuthFixture(t)
	uid := "user-change-email-ok"
	fixture.addAccount(uid, "old@example.com")
	ctx := actor.WithUser(context.Background(), uid)

	// 只有 change_email 用途的码能改邮箱。
	fixture.emailCodes.codes["new@example.com"] = fixture.authenticator.digest("123456")
	fixture.emailCodes.purposes["new@example.com"] = "login"
	fixture.solveCaptcha(t, "cap-2", "8")
	if _, err := fixture.authenticator.ChangeEmail(ctx, "new@example.com", "123456", "cap-2", "8"); err == nil {
		t.Fatal("login code must not change the email")
	}

	fixture.emailCodes.codes["new@example.com"] = fixture.authenticator.digest("654321")
	fixture.emailCodes.purposes["new@example.com"] = "change_email"
	if _, err := fixture.authenticator.ChangeEmail(ctx, "new@example.com", "654321", "cap-2", "8"); err != nil {
		t.Fatalf("ChangeEmail: %v", err)
	}
	if got := fixture.users.byUID[uid].Email; got != "new@example.com" {
		t.Fatalf("email = %q, want new@example.com", got)
	}
}

// TestChangeEmailRejectsAddressOwnedByAnotherAccount 不能把邮箱改成别人已经在用的地址。
func TestChangeEmailRejectsAddressOwnedByAnotherAccount(t *testing.T) {
	fixture := newAuthFixture(t)
	fixture.addAccount("user-a", "a@example.com")
	fixture.addAccount("user-b", "b@example.com")
	ctx := actor.WithUser(context.Background(), "user-a")

	fixture.emailCodes.codes["b@example.com"] = fixture.authenticator.digest("111222")
	fixture.emailCodes.purposes["b@example.com"] = "change_email"
	fixture.solveCaptcha(t, "cap-3", "5")
	err := func() error {
		_, err := fixture.authenticator.ChangeEmail(ctx, "b@example.com", "111222", "cap-3", "5")
		return err
	}()
	assertClientMessage(t, err, "email already exists")
}

// TestSendSelfChangeEmailCodeRequiresLogin 未登录时不能给任意地址发码。
func TestSendSelfChangeEmailCodeRequiresLogin(t *testing.T) {
	fixture := newAuthFixture(t)
	err := fixture.authenticator.SendSelfChangeEmailCode(context.Background(), "new@example.com", "cap", "1")
	if err == nil {
		t.Fatal("unauthenticated send must fail")
	}
}

// TestChangeEmailIsRateLimitedByAttempts 新地址的码同样受尝试次数限制：五次错误后作废。
func TestChangeEmailIsRateLimitedByAttempts(t *testing.T) {
	fixture := newAuthFixture(t)
	uid := "user-change-email-attempts"
	fixture.addAccount(uid, "old@example.com")
	ctx := actor.WithUser(context.Background(), uid)
	fixture.emailCodes.codes["target@example.com"] = fixture.authenticator.digest("999999")
	fixture.emailCodes.purposes["target@example.com"] = "change_email"
	fixture.solveCaptcha(t, "cap-4", "6")

	for i := 0; i < emailCodeMaxAttempts; i++ {
		_, _ = fixture.authenticator.ChangeEmail(ctx, "target@example.com", "000000", "cap-4", "6")
	}
	if got := fixture.emailCodes.exhausted["target@example.com"]; got != 1 {
		t.Fatalf("exhausted = %d, want 1", got)
	}
}
