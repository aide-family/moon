package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gopkg.in/gomail.v2"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/conf"
)

func NewEmailRepository(c *conf.Bootstrap) repository.Email {
	emailConfig := c.GetGlobalEmail()
	return &emailRepositoryImpl{dialer: gomail.NewDialer(emailConfig.Host, int(emailConfig.Port), emailConfig.Username, emailConfig.Password)}
}

type emailRepositoryImpl struct {
	dialer *gomail.Dialer
}

// SendEmail implements [repository.Email].
func (e *emailRepositoryImpl) SendEmail(ctx context.Context, req *bo.SendEmailBo) error {
	log.Context(ctx).Debugw("msg", "send email", "params", req)
	msg := gomail.NewMessage(gomail.SetCharset("UTF-8"), gomail.SetEncoding(gomail.Base64))
	msg.SetHeader("From", e.dialer.Username)
	msg.SetHeader("To", req.To...)
	msg.SetHeader("Cc", req.Cc...)
	msg.SetHeader("Subject", req.Subject)
	msg.SetBody(req.ContentType, req.Body)
	for key, values := range req.Headers {
		msg.SetHeader(key, values...)
	}
	return e.dialer.DialAndSend(msg)
}
