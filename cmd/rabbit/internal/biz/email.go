package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/biz/repository"
	"github.com/moon-monitor/moon/pkg/merr"
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
