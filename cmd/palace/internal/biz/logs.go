package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
)

func NewLogsBiz(
	sendMessageLogRepo repository.SendMessageLog,
	operateLogRepo repository.OperateLog,
	transaction repository.Transaction,
	logger log.Logger,
) *Logs {
	return &Logs{
		sendMessageLogRepo: sendMessageLogRepo,
		operateLogRepo:     operateLogRepo,
		transaction:        transaction,
		helper:             log.NewHelper(log.With(logger, "module", "biz.logs")),
	}
}

type Logs struct {
	sendMessageLogRepo repository.SendMessageLog
	operateLogRepo     repository.OperateLog
	transaction        repository.Transaction
	helper             *log.Helper
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

func (l *Logs) CreateOperateLog(ctx context.Context, params *bo.OperateLogParams) {
	if params.TeamID > 0 {
		if err := l.operateLogRepo.TeamCreateLog(ctx, params); err != nil {
			l.helper.WithContext(ctx).Warnw("msg", "create team operate log failed", "err", err)
		}
		return
	}

	if err := l.operateLogRepo.CreateLog(ctx, params); err != nil {
		l.helper.WithContext(ctx).Warnw("msg", "create operate log failed", "err", err)
	}
}
