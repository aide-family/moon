package event

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

var _ do.SendMessageLog = (*SendMessageLog)(nil)

const tableNameSendMessageLog = "team_send_message_logs"

type SendMessageLog struct {
	do.BaseModel
	TeamID      uint32                 `gorm:"column:team_id;type:int unsigned;not null;comment:团队ID" json:"team_id,omitempty"`
	SentAt      time.Time              `gorm:"column:sent_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:发送时间" json:"sent_at,omitempty"`
	MessageType vobj.MessageType       `gorm:"column:message_type;type:tinyint(2);not null;comment:消息类型" json:"message_type,omitempty"`
	Message     string                 `gorm:"column:message;type:text;not null;comment:消息内容" json:"message,omitempty"`
	RequestID   string                 `gorm:"column:request_id;type:varchar(64);not null;comment:请求ID;uniqueIndex:idx__request_id" json:"request_id,omitempty"`
	Status      vobj.SendMessageStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status,omitempty"`
	RetryCount  int                    `gorm:"column:retry_count;type:int unsigned;not null;comment:重试次数" json:"retry_count,omitempty"`
	Error       string                 `gorm:"column:error;type:text;not null;comment:错误信息" json:"error,omitempty"`
}

// GetError implements do.SendMessageLog.
func (s *SendMessageLog) GetError() string {
	if s == nil {
		return ""
	}
	return s.Error
}

// GetMessage implements do.SendMessageLog.
func (s *SendMessageLog) GetMessage() string {
	if s == nil {
		return ""
	}
	return s.Message
}

// GetMessageType implements do.SendMessageLog.
func (s *SendMessageLog) GetMessageType() vobj.MessageType {
	if s == nil {
		return vobj.MessageTypeUnknown
	}
	return s.MessageType
}

// GetRequestID implements do.SendMessageLog.
func (s *SendMessageLog) GetRequestID() string {
	if s == nil {
		return ""
	}
	return s.RequestID
}

// GetRetryCount implements do.SendMessageLog.
func (s *SendMessageLog) GetRetryCount() int32 {
	if s == nil {
		return 0
	}
	return int32(s.RetryCount)
}

// GetStatus implements do.SendMessageLog.
func (s *SendMessageLog) GetStatus() vobj.SendMessageStatus {
	if s == nil {
		return vobj.SendMessageStatusUnknown
	}
	return s.Status
}

// GetTeamID implements do.SendMessageLog.
func (s *SendMessageLog) GetTeamID() uint32 {
	if s == nil {
		return 0
	}
	return s.TeamID
}

func (s *SendMessageLog) TableName() string {
	return genSendMessageLogTableName(s.TeamID, s.SentAt)
}

func createSendMessageLogTable(teamId uint32, t time.Time, tx *gorm.DB) (err error) {
	tableName := genSendMessageLogTableName(teamId, t)
	if do.HasTable(teamId, tx, tableName) {
		return
	}
	s := &SendMessageLog{
		TeamID: teamId,
		SentAt: t,
	}
	if err := do.CreateTable(teamId, tx, tableName, s); err != nil {
		return err
	}
	return
}

func genSendMessageLogTableName(teamId uint32, t time.Time) string {
	weekStart := do.GetPreviousMonday(t)

	return fmt.Sprintf("%s_%d_%s", tableNameSendMessageLog, teamId, weekStart.Format("20060102"))
}

func GetSendMessageLogTableName(teamId uint32, t time.Time, tx *gorm.DB) (string, error) {
	tableName := genSendMessageLogTableName(teamId, t)
	if !do.HasTable(teamId, tx, tableName) {
		return tableName, createSendMessageLogTable(teamId, t, tx)
	}
	return tableName, nil
}

func GetSendMessageLogTableNames(teamId uint32, start, end time.Time, tx *gorm.DB) []string {
	if start.After(end) {
		return nil
	}

	var tableNames []string
	// 找到第一个周一（包含或早于start的周一）
	firstMonday := do.GetPreviousMonday(start)

	// 从第一个周一开始，每周增加7天，直到超过end时间
	for currentMonday := firstMonday; !currentMonday.After(end); currentMonday = currentMonday.AddDate(0, 0, 7) {
		// 确保生成的表名在时间范围内（周一+6天不超过start）
		if currentMonday.AddDate(0, 0, 6).Before(start) {
			continue
		}
		if do.HasTable(teamId, tx, genSendMessageLogTableName(teamId, currentMonday)) {
			tableNames = append(tableNames, genSendMessageLogTableName(teamId, currentMonday))
		}
	}

	return tableNames
}
