package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
)

func NewLogs(
	sendMessageLogRepo repository.SendMessageLog,
	transaction repository.Transaction,
) *Logs {
	return &Logs{
		sendMessageLogRepo: sendMessageLogRepo,
		transaction:        transaction,
	}
}

type Logs struct {
	sendMessageLogRepo repository.SendMessageLog
	transaction        repository.Transaction
}

func (l *Logs) GetSendMessageLogs(ctx context.Context, params *bo.ListSendMessageLogParams) (*bo.ListSendMessageLogReply, error) {
	return l.sendMessageLogRepo.List(ctx, params)
}

func (l *Logs) GetSendMessageLog(ctx context.Context, params *bo.GetSendMessageLogParams) (do.SendMessageLog, error) {
	return l.sendMessageLogRepo.Get(ctx, params)
}

func (l *Logs) UpdateSendMessageLogStatus(ctx context.Context, params *bo.UpdateSendMessageLogStatusParams) error {
	return l.sendMessageLogRepo.UpdateStatus(ctx, params)
}

func (l *Logs) RetrySendMessage(ctx context.Context, params *bo.RetrySendMessageParams) error {
	req := &bo.GetSendMessageLogParams{
		TeamID:    params.TeamID,
		RequestID: params.RequestID,
	}
	sendMessageLog, err := l.sendMessageLogRepo.Get(ctx, req)
	if err != nil {
		return err
	}
	if !sendMessageLog.GetStatus().IsFailed() {
		return merr.ErrorParams("message is %s, do not need retry", sendMessageLog.GetStatus())
	}
	if params.TeamID > 0 {
		return l.transaction.BizExec(ctx, func(ctx context.Context) error {
			if err := l.sendMessage(ctx, sendMessageLog); err != nil {
				return err
			}
			return l.sendMessageLogRepo.Retry(ctx, params)
		})
	}

	return l.transaction.MainExec(ctx, func(ctx context.Context) error {
		if err := l.sendMessage(ctx, sendMessageLog); err != nil {
			return err
		}
		return l.sendMessageLogRepo.Retry(ctx, params)
	})
}

func (l *Logs) sendMessage(ctx context.Context, sendMessageLog do.SendMessageLog) error {
	return nil
}
