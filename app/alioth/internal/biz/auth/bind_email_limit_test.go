package auth

import (
	"context"
	"fmt"
	"testing"
)

// TestBindEmailCodeLimitedPerBindingToken 一个第三方绑定凭证最多给 bindEmailLimit 个
// 不同邮箱发码；超过后拒绝，避免 bindToken 被当成邮件轰炸入口。
func TestBindEmailCodeLimitedPerBindingToken(t *testing.T) {
	fixture := newAuthFixture(t)
	binding, err := fixture.authenticator.newOAuthBinding("GITHUB", oauthProfile{openID: "oid-limit"}, "")
	if err != nil {
		t.Fatal(err)
	}
	token := binding.Token
	hash := fixture.authenticator.tokenHash(token)
	for i := 0; i < bindEmailLimit; i++ {
		fixture.emailCodes.bindSends = append(fixture.emailCodes.bindSends, bindSend{
			tokenHash: hash,
			email:     fmt.Sprintf("addr%d@example.com", i),
		})
	}
	fixture.solveCaptcha(t, "cap-bind-limit", "3")

	err = fixture.authenticator.SendOAuthBindEmailCode(context.Background(), token, "new@example.com", "cap-bind-limit", "3")
	assertClientMessage(t, err, "too many bind email addresses")
}

// TestBindEmailCodeAllowsOtherTokens 不同绑定凭证各自计数，互不影响。
func TestBindEmailCodeAllowsOtherTokens(t *testing.T) {
	fixture := newAuthFixture(t)
	exhaustedToken, err := fixture.authenticator.newOAuthBinding("GITHUB", oauthProfile{openID: "oid-a"}, "")
	if err != nil {
		t.Fatal(err)
	}
	hash := fixture.authenticator.tokenHash(exhaustedToken.Token)
	for i := 0; i < bindEmailLimit; i++ {
		fixture.emailCodes.bindSends = append(fixture.emailCodes.bindSends, bindSend{
			tokenHash: hash,
			email:     fmt.Sprintf("addr%d@example.com", i),
		})
	}
	otherToken, err := fixture.authenticator.newOAuthBinding("GITHUB", oauthProfile{openID: "oid-b"}, "")
	if err != nil {
		t.Fatal(err)
	}
	fixture.solveCaptcha(t, "cap-bind-other", "4")

	if err := fixture.authenticator.SendOAuthBindEmailCode(context.Background(), otherToken.Token, "other@example.com", "cap-bind-other", "4"); err != nil {
		t.Fatalf("another binding token must not be limited: %v", err)
	}
	if len(fixture.mail.recipients) == 0 {
		t.Fatal("expected the bind code mail to be sent")
	}
}
