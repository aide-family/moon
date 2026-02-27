package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToMessageLogItemBo(messageLogDo *do.MessageLog) *bo.MessageLogItemBo {
	return &bo.MessageLogItemBo{
		NamespaceUID: messageLogDo.NamespaceUID,
		UID:          messageLogDo.ID,
		SendAt:       messageLogDo.SendAt,
		Message:      messageLogDo.Message,
		Config:       messageLogDo.Config,
		MessageType:  messageLogDo.Type,
		Status:       messageLogDo.Status,
		RetryTotal:   messageLogDo.RetryTotal,
		LastError:    messageLogDo.LastError,
		CreatedAt:    messageLogDo.CreatedAt,
		UpdatedAt:    messageLogDo.UpdatedAt,
	}
}

func ToMessageLogDo(ctx context.Context, messageLog *bo.CreateMessageLogBo) *do.MessageLog {
	model := &do.MessageLog{
		NamespaceUID: contextx.GetNamespace(ctx),
		Message:      messageLog.Message,
		Config:       messageLog.Config,
		Type:         messageLog.MessageType,
		Status:       enum.MessageStatus_PENDING,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	return model
}
