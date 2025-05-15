package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
)

type Send interface {
	Email(ctx context.Context, params bo.SendEmailParams) error
	SMS(ctx context.Context, params bo.SendSMSParams) error
	Hook(ctx context.Context, params bo.SendHookParams) error
}
