package email

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

var _ Email = (*e)(nil)

type (
	// Email email
	e struct {
		config Config
		mail   *gomail.Message
	}

	// Email email interface
	Email interface {
		Send() error
		SetTo(to ...string) Email
		SetSubject(subject string) Email
		SetBody(body string, contentType ...string) Email
		SetAttach(attach ...string) Email
		SetCc(cc ...string) Email
		Copy() Email
	}

	// Config email configuration
	Config interface {
		GetUser() string
		GetPass() string
		GetHost() string
		GetPort() uint32
		GetEnable() bool
		GetName() string
	}
)

const (
	// DOMAIN domain name
	DOMAIN = "Moon Monitoring System"
)

func (l *e) Copy() Email {
	return &e{
		config: copyConfig(l.config),
		mail:   gomail.NewMessage(gomail.SetEncoding(gomail.Base64)),
	}
}

// init initialize
func (l *e) init() Email {
	if l.mail == nil {
		l.mail = gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	}
	return l
}

// SetTo set recipients
func (l *e) SetTo(to ...string) Email {
	l.mail.SetHeader("To", to...) // send to users (can be multiple)
	return l
}

// SetCc set CC recipients
func (l *e) SetCc(cc ...string) Email {
	l.mail.SetHeader("Cc", cc...)
	return l
}

// SetSubject set email subject
func (l *e) SetSubject(subject string) Email {
	l.mail.SetHeader("Subject", subject) // set email subject
	return l
}

// SetBody set email body
func (l *e) SetBody(body string, contentType ...string) Email {
	cType := "text/plain"
	if len(contentType) > 0 && contentType[0] != "" {
		cType = contentType[0]
	}

	l.mail.SetBody(cType, body) // set email body
	return l
}

// SetAttach set attachments
func (l *e) SetAttach(attach ...string) Email {
	for _, v := range attach {
		l.mail.Attach(v)
	}
	return l
}

// setFrom set sender
func (l *e) setFrom(from string) Email {
	domain := DOMAIN
	if from != "" {
		domain = from
	}

	l.mail.SetHeader("From", l.mail.FormatAddress(l.config.GetUser(), domain)) // add alias
	return l
}

// Send send email
func (l *e) Send() error {
	l.setFrom(l.config.GetUser())
	d := gomail.NewDialer(l.config.GetHost(), int(l.config.GetPort()), l.config.GetUser(), l.config.GetPass()) // set email body
	d.TLSConfig = &tls.Config{InsecureSkipVerify: false, ServerName: l.config.GetHost(), MinVersion: tls.VersionTLS12}
	err := d.DialAndSend(l.mail)
	return err
}

// New create email
func New(cfg Config) Email {
	if !cfg.GetEnable() {
		return NewMockEmail()
	}
	instance := &e{config: cfg}
	return instance.init()
}
