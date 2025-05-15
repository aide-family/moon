package email

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

var _ Email = (*e)(nil)

type (
	// Email 邮件
	e struct {
		config Config
		mail   *gomail.Message
	}

	// Email 邮件接口
	Email interface {
		Send() error
		SetTo(to ...string) Email
		SetSubject(subject string) Email
		SetBody(body string, contentType ...string) Email
		SetAttach(attach ...string) Email
		SetCc(cc ...string) Email
		Copy() Email
	}

	// Config 邮件配置
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
	// DOMAIN 域名
	DOMAIN = "Moon监控系统"
)

func (l *e) Copy() Email {
	return &e{
		config: copyConfig(l.config),
		mail:   gomail.NewMessage(gomail.SetEncoding(gomail.Base64)),
	}
}

// init 初始化
func (l *e) init() Email {
	if l.mail == nil {
		l.mail = gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	}
	return l
}

// SetTo 设置收件人
func (l *e) SetTo(to ...string) Email {
	l.mail.SetHeader("To", to...) // 发送给用户(可以多个)
	return l
}

// SetCc 设置抄送人
func (l *e) SetCc(cc ...string) Email {
	l.mail.SetHeader("Cc", cc...)
	return l
}

// SetSubject 设置邮件主题
func (l *e) SetSubject(subject string) Email {
	l.mail.SetHeader("Subject", subject) // 设置邮件主题
	return l
}

// SetBody 设置邮件正文
func (l *e) SetBody(body string, contentType ...string) Email {
	cType := "text/plain"
	if len(contentType) > 0 && contentType[0] != "" {
		cType = contentType[0]
	}

	l.mail.SetBody(cType, body) // 设置邮件正文
	return l
}

// SetAttach 设置附件
func (l *e) SetAttach(attach ...string) Email {
	for _, v := range attach {
		l.mail.Attach(v)
	}
	return l
}

// setFrom 设置发件人
func (l *e) setFrom(from string) Email {
	domain := DOMAIN
	if from != "" {
		domain = from
	}

	l.mail.SetHeader("From", l.mail.FormatAddress(l.config.GetUser(), domain)) // 添加别名
	return l
}

// Send 发送邮件
func (l *e) Send() error {
	l.setFrom(l.config.GetUser())
	d := gomail.NewDialer(l.config.GetHost(), int(l.config.GetPort()), l.config.GetUser(), l.config.GetPass()) // 设置邮件正文
	d.TLSConfig = &tls.Config{InsecureSkipVerify: false, ServerName: l.config.GetHost(), MinVersion: tls.VersionTLS12}
	err := d.DialAndSend(l.mail)
	return err
}

// New 创建邮件
func New(cfg Config) Email {
	if !cfg.GetEnable() {
		return NewMockEmail()
	}
	instance := &e{config: cfg}
	return instance.init()
}
