package mailer

import (
	"log/slog"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/pkg/merr"
)

func newTestSender(from, username string) *Sender {
	return NewSender(&conf.Auth{Smtp: &conf.Auth_SMTP{
		Host:     "smtp.example.com",
		Port:     25,
		Username: username,
		Password: "secret",
		From:     from,
	}}, slog.Default())
}

// TestSenderIdentity 覆盖三种 From 写法：裸地址、带显示名、以及历史配置里只写显示名。
// 最后一种要退化为 SMTP 用户名当信封地址，否则 MAIL FROM 里的空格和中文会被服务端拒绝。
func TestSenderIdentity(t *testing.T) {
	cases := []struct {
		name         string
		from         string
		username     string
		wantEnvelope string
	}{
		{name: "bare address", from: "aidedevops@163.com", wantEnvelope: "aidedevops@163.com"},
		{name: "display name and address", from: "Alioth 瑶光 <aidedevops@163.com>", wantEnvelope: "aidedevops@163.com"},
		{name: "display name only falls back to username", from: "Alioth 瑶光", username: "aidedevops@163.com", wantEnvelope: "aidedevops@163.com"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			sender := newTestSender(testCase.from, testCase.username)
			envelope, header, err := sender.senderIdentity()
			if err != nil {
				t.Fatalf("senderIdentity: %v", err)
			}
			if envelope != testCase.wantEnvelope {
				t.Fatalf("信封发件人 = %q, want %q", envelope, testCase.wantEnvelope)
			}
			if !strings.Contains(header, testCase.wantEnvelope) {
				t.Fatalf("From 头 = %q，应包含地址 %q", header, testCase.wantEnvelope)
			}
			if strings.Contains(header, "Alioth 瑶光") {
				t.Fatalf("From 头里的显示名没有被 RFC 2047 编码：%q", header)
			}
		})
	}
}

// TestSenderIdentityRejectsDisplayNameWithoutUsername 保证配置只写了显示名、又没有用户名时
// 返回明确的服务端配置错误，而不是把非法信封地址发给 SMTP 再收到 500。
func TestSenderIdentityRejectsDisplayNameWithoutUsername(t *testing.T) {
	sender := newTestSender("Alioth 瑶光", "")
	if _, _, err := sender.senderIdentity(); err == nil {
		t.Fatal("senderIdentity 应返回配置错误")
	} else if kerrors.Reason(err) != merr.ReasonServer {
		t.Fatalf("错误类型 = %v, want server error", err)
	}
}

func TestRecipientAddress(t *testing.T) {
	for _, testCase := range []struct{ in, want string }{
		{in: " user@example.com ", want: "user@example.com"},
		{in: "User <user@example.com>", want: "user@example.com"},
	} {
		got, err := recipientAddress(testCase.in)
		if err != nil {
			t.Fatalf("recipientAddress(%q): %v", testCase.in, err)
		}
		if got != testCase.want {
			t.Fatalf("recipientAddress(%q) = %q, want %q", testCase.in, got, testCase.want)
		}
	}
	for _, in := range []string{"not an email", "user@example.com\r\nBcc: other@example.com"} {
		if _, err := recipientAddress(in); err == nil {
			t.Fatalf("recipientAddress(%q) 应返回错误", in)
		} else if kerrors.Reason(err) != merr.ReasonClient {
			t.Fatalf("recipientAddress(%q) 错误类型 = %v, want client error", in, err)
		}
	}
}

// TestSendRejectsHeaderInjection 只验证协议之前的校验，不会真的连 SMTP。
func TestSendRejectsHeaderInjection(t *testing.T) {
	sender := newTestSender("aidedevops@163.com", "")
	if err := sender.Send("user@example.com", "subject\r\nBcc: other@example.com", "body"); err == nil {
		t.Fatal("主题里的换行应被拒绝")
	} else if kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("主题注入错误类型 = %v, want client error", err)
	}
	if err := sender.Send("user@example.com\r\nBcc: other@example.com", "subject", "body"); err == nil {
		t.Fatal("收件人里的换行应被拒绝")
	}
}

// TestBuildMessage 校验头部、空行和 base64 正文的结构。
func TestBuildMessage(t *testing.T) {
	message := string(buildMessage("Alioth <aidedevops@163.com>", "user@example.com", "Alioth bind code", "Your code is 123456\r\n"))
	for _, want := range []string{
		"From: Alioth <aidedevops@163.com>\r\n",
		"To: user@example.com\r\n",
		"Subject: Alioth bind code\r\n",
		"Content-Type: text/plain; charset=UTF-8\r\n",
		"\r\n\r\n",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("邮件缺少 %q：\n%s", want, message)
		}
	}
	if strings.Contains(message, "Your code is 123456") {
		t.Fatal("正文应使用 base64 编码")
	}
}
