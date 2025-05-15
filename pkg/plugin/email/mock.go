package email

import (
	"github.com/go-kratos/kratos/v2/log"
)

type mockEmail struct{}

func (m *mockEmail) Copy() Email {
	return m
}

// Send 发送邮件
func (m *mockEmail) Send() error {
	return nil
}

// SetTo 设置收件人
func (m *mockEmail) SetTo(to ...string) Email {
	log.Debugw("method", "SetTo", "to", to)
	return m
}

// SetSubject 设置邮件主题
func (m *mockEmail) SetSubject(subject string) Email {
	log.Debugw("method", "SetSubject", "subject", subject)
	return m
}

// SetBody 设置邮件正文
func (m *mockEmail) SetBody(body string, contentType ...string) Email {
	log.Debugw("method", "SetBody", "body", body, "contentType", contentType)
	return m
}

// SetAttach 设置附件
func (m *mockEmail) SetAttach(attach ...string) Email {
	log.Debugw("method", "SetAttach", "attach", attach)
	return m
}

// SetCc 设置抄送人
func (m *mockEmail) SetCc(cc ...string) Email {
	log.Debugw("method", "SetCc", "cc", cc)
	return m
}

// NewMockEmail 创建邮件模拟
func NewMockEmail() Email {
	return &mockEmail{}
}
