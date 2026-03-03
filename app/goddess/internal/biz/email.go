package biz

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"

	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/magicbox/merr"
	"github.com/go-kratos/kratos/v2/log"
)

func NewEmail(c *conf.Bootstrap) *Email {
	emailConfig := c.GetGlobalEmail()
	return &Email{dialer: gomail.NewDialer(emailConfig.Host, int(emailConfig.Port), emailConfig.Username, emailConfig.Password)}
}

type Email struct {
	dialer *gomail.Dialer
}

func (e *Email) SendEmailLoginCode(ctx context.Context, email, codeID, code string) error {
	log.Context(ctx).Debugw("msg", "send email login code", "email", email, "codeID", codeID, "code", code)
	msg := gomail.NewMessage(gomail.SetCharset("UTF-8"), gomail.SetEncoding(gomail.Base64))
	msg.SetHeader("From", e.dialer.Username)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Moon Goddess Email Login Code")
	msg.SetBody("text/html", fmt.Sprintf("Your email login code is <b>%s</b>. Please use it to login to Moon Goddess.", code))
	if err := e.dialer.DialAndSend(msg); err != nil {
		return merr.ErrorInternalServer("send email login code failed").WithCause(err)
	}
	return nil
}
