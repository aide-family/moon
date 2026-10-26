package auth

import (
	"context"
	"testing"
)

// TestEmailLoginExhaustsCodeAfterFiveFailures 同一个码错 5 次后立即作废，
// 即使随后填对也不能登录；作废次数进入邮箱级计数。
func TestEmailLoginExhaustsCodeAfterFiveFailures(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "exhaust@example.com"
	fixture.addUser(email)
	fixture.emailCodes.codes[email] = fixture.authenticator.digest("111111")
	fixture.emailCodes.purposes[email] = "login"
	fixture.solveCaptcha(t, "cap-exhaust", "7")

	for i := 0; i < emailCodeMaxAttempts; i++ {
		_, err := fixture.authenticator.EmailLogin(context.Background(), email, "000000", "cap-exhaust", "7")
		assertClientMessage(t, err, "invalid email or code")
	}
	if got := fixture.emailCodes.exhausted[email]; got != 1 {
		t.Fatalf("exhausted = %d, want 1", got)
	}
	// 码已作废：正确答案也被拒绝（此时失败次数 1 < 锁定阈值，仍是普通错误）。
	_, err := fixture.authenticator.EmailLogin(context.Background(), email, "111111", "cap-exhaust", "7")
	assertClientMessage(t, err, "invalid email or code")
}

// TestEmailLoginLockedOutAfterRepeatedFailures 窗口内作废次数达到上限后，
// 即使验证码正确也不签发会话，必须等锁定窗口滑过。
func TestEmailLoginLockedOutAfterRepeatedFailures(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "locked@example.com"
	fixture.addUser(email)
	fixture.emailCodes.exhausted[email] = emailCodeFailureLimit
	fixture.solveCaptcha(t, "cap-lockout", "7")

	_, err := fixture.authenticator.EmailLogin(context.Background(), email, "123456", "cap-lockout", "7")
	assertClientMessage(t, err, "too many verification attempts")
}

// TestSendLoginCodeRefusedWhileLockedOut 锁定期间连发码也拒绝，避免攻击者靠重发刷新可猜次数。
func TestSendLoginCodeRefusedWhileLockedOut(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "locked2@example.com"
	fixture.addUser(email)
	fixture.emailCodes.exhausted[email] = emailCodeFailureLimit
	fixture.solveCaptcha(t, "cap-lockout2", "9")

	err := fixture.authenticator.SendEmailLoginCode(context.Background(), email, "cap-lockout2", "9")
	assertClientMessage(t, err, "too many verification attempts")
}
