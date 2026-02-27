package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type MessageRetryLog struct {
	BaseModel

	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;index:idx__message_retry_log__namespace_uid"`
	MessageLogID snowflake.ID `gorm:"column:message_log_id;index:idx__message_retry_log__message_log_id"`
	RetryAt      time.Time    `gorm:"column:retry_at"`
}

func (MessageRetryLog) TableName() string {
	return "message_retry_logs"
}
