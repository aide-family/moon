package repository

import (
	"context"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

type MessageLog interface {
	CreateMessageLog(ctx context.Context, messageLog *bo.CreateMessageLogBo) (snowflake.ID, error)
	ListMessageLog(ctx context.Context, req *bo.ListMessageLogBo) (*bo.PageResponseBo[*bo.MessageLogItemBo], error)
	GetMessageLog(ctx context.Context, uid snowflake.ID) (*bo.MessageLogItemBo, error)
	GetAllMessageLogs(ctx context.Context, status enum.MessageStatus) ([]*bo.MessageLogItemBo, error)
	GetMessageLogWithLock(ctx context.Context, uid snowflake.ID) (*bo.MessageLogItemBo, error)
	UpdateMessageLogStatusIf(ctx context.Context, uid snowflake.ID, oldStatus, newStatus enum.MessageStatus) (bool, error)
	UpdateMessageLogStatusSendingIf(ctx context.Context, uid snowflake.ID, oldStatus enum.MessageStatus) (bool, error)
	UpdateMessageLogLastErrorIf(ctx context.Context, uid snowflake.ID, oldStatus enum.MessageStatus, lastError string) (bool, error)
	UpdateMessageLogStatusSuccessIf(ctx context.Context, uid snowflake.ID) (bool, error)
	MessageLogRetryIncrement(ctx context.Context, uid snowflake.ID) error
}

type MessageRetryLog interface {
	CreateMessageRetryLog(ctx context.Context, msg *bo.MessageLogItemBo) error
}
