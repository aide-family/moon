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
	TeamID      uint32                 `gorm:"column:team_id;type:int unsigned;not null;comment:team ID" json:"team_id,omitempty"`
	SentAt      time.Time              `gorm:"column:sent_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:sent time" json:"sent_at,omitempty"`
	MessageType vobj.MessageType       `gorm:"column:message_type;type:tinyint(2);not null;comment:message type" json:"message_type,omitempty"`
	Message     string                 `gorm:"column:message;type:text;not null;comment:message content" json:"message,omitempty"`
	RequestID   string                 `gorm:"column:request_id;type:varchar(64);not null;comment:request ID;uniqueIndex:idx__request_id" json:"request_id,omitempty"`
	Status      vobj.SendMessageStatus `gorm:"column:status;type:tinyint(2);not null;comment:status" json:"status,omitempty"`
	RetryCount  int                    `gorm:"column:retry_count;type:int unsigned;not null;comment:retry count" json:"retry_count,omitempty"`
	Error       string                 `gorm:"column:error;type:text;not null;comment:error message" json:"error,omitempty"`
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

func createSendMessageLogTable(teamID uint32, t time.Time, tx *gorm.DB) (err error) {
	tableName := genSendMessageLogTableName(teamID, t)
	if do.HasTable(teamID, tx, tableName) {
		return
	}
	s := &SendMessageLog{
		TeamID: teamID,
		SentAt: t,
	}
	if err := do.CreateTable(teamID, tx, tableName, s); err != nil {
		return err
	}
	return
}

func genSendMessageLogTableName(teamID uint32, t time.Time) string {
	weekStart := do.GetPreviousMonday(t)

	return fmt.Sprintf("%s_%d_%s", tableNameSendMessageLog, teamID, weekStart.Format("20060102"))
}

func GetSendMessageLogTableName(teamID uint32, t time.Time, tx *gorm.DB) (string, error) {
	tableName := genSendMessageLogTableName(teamID, t)
	if !do.HasTable(teamID, tx, tableName) {
		return tableName, createSendMessageLogTable(teamID, t, tx)
	}
	return tableName, nil
}

func GetSendMessageLogTableNames(teamID uint32, start, end time.Time, tx *gorm.DB) []string {
	if start.After(end) {
		return nil
	}

	var tableNames []string
	// Find the first Monday (Monday containing or before start)
	firstMonday := do.GetPreviousMonday(start)

	// From the first Monday, add 7 days each week until exceeding end time
	for currentMonday := firstMonday; !currentMonday.After(end); currentMonday = currentMonday.AddDate(0, 0, 7) {
		// Ensure the generated table name is within the time range (Monday + 6 days not before start)
		if currentMonday.AddDate(0, 0, 6).Before(start) {
			continue
		}
		if do.HasTable(teamID, tx, genSendMessageLogTableName(teamID, currentMonday)) {
			tableNames = append(tableNames, genSendMessageLogTableName(teamID, currentMonday))
		}
	}

	return tableNames
}
