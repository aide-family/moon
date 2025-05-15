package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type SendMessageLog interface {
	Get(ctx context.Context, params *bo.GetSendMessageLogParams) (do.SendMessageLog, error)
	Create(ctx context.Context, params *bo.CreateSendMessageLogParams) error
	UpdateStatus(ctx context.Context, params *bo.UpdateSendMessageLogStatusParams) error
	List(ctx context.Context, params *bo.ListSendMessageLogParams) (*bo.ListSendMessageLogReply, error)
	Retry(ctx context.Context, params *bo.RetrySendMessageParams) error
}

type SendMessage interface {
	SendEmail(ctx context.Context, params *bo.SendEmailParams) error
}
