package email

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

var _ Interface = (*Email)(nil)

type (
	// Email 邮件
	Email struct {
		config Config
		mail   *gomail.Message
	}

	Interface interface {
		// Send 发送邮件
		Send() error
		SetTo(to ...string) Interface
		SetSubject(subject string) Interface
		SetBody(body string, contentType ...string) Interface
		SetAttach(attach ...string) Interface
		SetCc(cc ...string) Interface
	}

	Config interface {
		GetUser() string
		GetPass() string
		GetHost() string
		GetPort() uint32
	}

	config struct {
		user string
		pass string
		host string
		port uint32
	}
)

func (c *config) GetUser() string {
	return c.user
}

func (c *config) GetPass() string {
	return c.pass
}

func (c *config) GetHost() string {
	return c.host
}

func (c *config) GetPort() uint32 {
	return c.port
}

const (
	DOMAIN = "Moon监控系统"
)

var _ Config = (*config)(nil)

func (l *Email) init() Interface {
	if l.mail == nil {
		l.mail = gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	}
	return l
}

func (l *Email) SetTo(to ...string) Interface {
	l.init()
	l.mail.SetHeader("To", to...) // 发送给用户(可以多个)
	return l
}

func (l *Email) SetCc(cc ...string) Interface {
	l.init()
	l.mail.SetHeader("Cc", cc...)
	return l
}

func (l *Email) SetSubject(subject string) Interface {
	l.init()
	l.mail.SetHeader("Subject", subject) // 设置邮件主题
	return l
}

func (l *Email) SetBody(body string, contentType ...string) Interface {
	cType := "text/plain"
	if len(contentType) > 0 && contentType[0] != "" {
		cType = contentType[0]
	}
	l.init()
	l.mail.SetBody(cType, body) // 设置邮件正文
	return l
}

func (l *Email) SetAttach(attach ...string) Interface {
	l.init()
	for _, v := range attach {
		l.mail.Attach(v)
	}
	return l
}

func (l *Email) setFrom(from string) *Email {
	domain := DOMAIN
	if from != "" {
		domain = from
	}

	l.init()
	l.mail.SetHeader("From", l.mail.FormatAddress(l.config.GetUser(), domain)) // 添加别名
	return l
}

func (l *Email) Send() error {
	l.init()
	l.setFrom(l.config.GetUser())
	/*
	   创建SMTP客户端，连接到远程的邮件服务器，需要指定服务器地址、端口号、用户名、密码，如果端口号为465的话，
	   自动开启SSL，这个时候需要指定TLSConfig
	*/
	d := gomail.NewDialer(l.config.GetHost(), int(l.config.GetPort()), l.config.GetUser(), l.config.GetPass()) // 设置邮件正文
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true, ServerName: l.config.GetHost()}
	err := d.DialAndSend(l.mail)
	return err
}

func New(cfg Config) Interface {
	return &Email{config: cfg}
}
