package email

import (
	"github.com/go-kratos/kratos/v2/log"
)

type mockEmail struct{}

func (m *mockEmail) Copy() Email {
	return m
}

// Send send email
func (m *mockEmail) Send() error {
	return nil
}

// SetTo set recipients
func (m *mockEmail) SetTo(to ...string) Email {
	log.Debugw("method", "SetTo", "to", to)
	return m
}

// SetSubject set email subject
func (m *mockEmail) SetSubject(subject string) Email {
	log.Debugw("method", "SetSubject", "subject", subject)
	return m
}

// SetBody set email body
func (m *mockEmail) SetBody(body string, contentType ...string) Email {
	log.Debugw("method", "SetBody", "body", body, "contentType", contentType)
	return m
}

// SetAttach set attachments
func (m *mockEmail) SetAttach(attach ...string) Email {
	log.Debugw("method", "SetAttach", "attach", attach)
	return m
}

// SetCc set CC recipients
func (m *mockEmail) SetCc(cc ...string) Email {
	log.Debugw("method", "SetCc", "cc", cc)
	return m
}

// NewMockEmail create email mock
func NewMockEmail() Email {
	return &mockEmail{}
}
