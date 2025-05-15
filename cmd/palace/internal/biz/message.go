package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/pkg/util/timex"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

func NewMessage(
	sendMessageRepo repository.SendMessage,
	sendMessageLogRepo repository.SendMessageLog,
	transaction repository.Transaction,
	logger log.Logger,
) *Message {
	return &Message{
		sendMessageRepo:    sendMessageRepo,
		sendMessageLogRepo: sendMessageLogRepo,
		helper:             log.NewHelper(log.With(logger, "module", "biz.message")),
		transaction:        transaction,
	}
}

type Message struct {
	sendMessageRepo    repository.SendMessage
	sendMessageLogRepo repository.SendMessageLog
	helper             *log.Helper
	transaction        repository.Transaction
}

func (a *Message) SendEmail(ctx context.Context, sendEmailParams *bo.SendEmailParams) error {
	sendMessageLogParams := &bo.CreateSendMessageLogParams{
		TeamID:      sendEmailParams.TeamID,
		SendAt:      timex.Now(),
		MessageType: vobj.MessageTypeEmail,
		Message:     sendEmailParams,
		RequestID:   sendEmailParams.RequestID,
	}
	transactionExecFun := a.transaction.MainExec
	if sendEmailParams.TeamID > 0 {
		transactionExecFun = a.transaction.BizExec
	}
	return transactionExecFun(ctx, func(ctx context.Context) error {
		if err := a.sendMessageLogRepo.Create(ctx, sendMessageLogParams); err != nil {
			a.helper.WithContext(ctx).Warnw("method", "create send message log error", "params", sendMessageLogParams, "error", err)
			return err
		}
		if err := a.sendMessageRepo.SendEmail(ctx, sendEmailParams); err != nil {
			a.helper.WithContext(ctx).Warnw("method", "send email error", "params", sendEmailParams, "error", err)
			return err
		}
		return nil
	})
}
