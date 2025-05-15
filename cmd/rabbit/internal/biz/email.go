package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/go-kratos/kratos/v2/log"
)

func NewEmail(sendRepo repository.Send, logger log.Logger) *Email {
	return &Email{
		helper:   log.NewHelper(log.With(logger, "module", "biz.email")),
		sendRepo: sendRepo,
	}
}

type Email struct {
	helper *log.Helper

	sendRepo repository.Send
}

func (e *Email) Send(ctx context.Context, params bo.SendEmailParams) error {
	if params.GetConfig() == nil {
		return merr.ErrorParamsError("No email configuration is available")
	}
	return e.sendRepo.Email(ctx, params)
}
