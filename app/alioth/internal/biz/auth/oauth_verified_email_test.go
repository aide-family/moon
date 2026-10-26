package auth

import (
	"context"
	"testing"
)

// TestOAuthVerifiedEmailReusesExistingAccount 平台已验证的第三方邮箱命中已有本地账号时，
// 直接登入该账号并绑定身份，不新建账号——这是"已验证邮箱可以并入已有账号"的核心安全前提。
func TestOAuthVerifiedEmailReusesExistingAccount(t *testing.T) {
	fixture := newAuthFixture(t)
	fixture.addAccount("user-known", "known@example.com")
	ctx := context.Background()

	account, err := fixture.authenticator.oauthAccountForVerifiedEmail(
		ctx,
		"GITHUB",
		oauthProfile{openID: "oid-known", email: "known@example.com", emailVerified: true, name: "known"},
		"known@example.com",
	)
	if err != nil {
		t.Fatalf("oauthAccountForVerifiedEmail: %v", err)
	}
	if account.UID != "user-known" {
		t.Fatalf("account.UID = %q, want user-known", account.UID)
	}
	link, ok := fixture.oauth.find("GITHUB", "oid-known")
	if !ok {
		t.Fatal("expected the OAuth identity to be linked to the existing account")
	}
	if link.userUID != "user-known" {
		t.Fatalf("linked user = %q, want user-known", link.userUID)
	}
}
