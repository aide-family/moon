package biz

import (
	"context"
	"fmt"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/magicbox/merr"
	"github.com/go-kratos/kratos/v2/log"
)

func NewEmail(emailRepo repository.Email) *Email {
	return &Email{emailRepo: emailRepo}
}

type Email struct {
	emailRepo repository.Email
}

func (e *Email) SendEmailLoginCode(ctx context.Context, email, codeID, code string) error {
	log.Context(ctx).Debugw("msg", "send email login code", "email", email, "codeID", codeID, "code", code)
	req := &bo.SendEmailBo{
		Subject:     "Moon Goddess Email Login Code",
		To:          []string{email},
		ContentType: "text/html",
		Body:        fmt.Sprintf("Your email login code is <b>%s</b>. Please use it to login to Moon Goddess.", code),
	}
	if err := e.emailRepo.SendEmail(ctx, req); err != nil {
		return merr.ErrorInternalServer("send email login code failed").WithCause(err)
	}
	return nil
}
