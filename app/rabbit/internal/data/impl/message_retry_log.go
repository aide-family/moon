package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/contextx"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/do"
	"github.com/aide-family/rabbit/internal/data/impl/query"
)

func NewMessageRetryLogRepository(d *data.Data) repository.MessageRetryLog {
	query.SetDefault(d.DB())
	return &messageRetryLogRepository{Data: d}
}

type messageRetryLogRepository struct {
	*data.Data
}

// CreateMessageRetryLog implements [repository.MessageRetryLog].
func (m *messageRetryLogRepository) CreateMessageRetryLog(ctx context.Context, msg *bo.MessageLogItemBo) error {
	messageRetryLog := &do.MessageRetryLog{
		NamespaceUID: msg.NamespaceUID,
		MessageLogID: msg.UID,
		RetryAt:      time.Now(),
	}
	messageRetryLog.WithCreator(contextx.GetUserUID(ctx))
	messageRetryLogMutation := query.Use(m.DB()).MessageRetryLog
	return messageRetryLogMutation.WithContext(ctx).Create(messageRetryLog)
}
