package do

import (
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

const (
	TableNameMessageLog = "message_logs"
)

type MessageLog struct {
	BaseModel

	NamespaceUID snowflake.ID          `gorm:"column:namespace_uid;index:idx__message_log__namespace_uid"`
	SendAt       time.Time             `gorm:"column:send_at;"`
	Message      strutil.EncryptString `gorm:"column:message;"`
	Config       strutil.EncryptString `gorm:"column:config;"`
	Type         enum.MessageType      `gorm:"column:type;default:0"`
	Status       enum.MessageStatus    `gorm:"column:status;default:0"`
	RetryTotal   int32                 `gorm:"column:retry_total;default:0"`
	LastError    string                `gorm:"column:last_error;"`
}

func (m *MessageLog) TableName() string {
	return TableNameMessageLog
}

func (m *MessageLog) BeforeCreate(tx *gorm.DB) (err error) {
	if err := m.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}
	if m.SendAt.IsZero() {
		m.SendAt = time.Now()
	}
	return nil
}

func GenMessageLogTableName(namespace snowflake.ID, sendAt time.Time) string {
	weekStart := getFirstMonday(sendAt)
	return strings.Join([]string{TableNameMessageLog, namespace.String(), weekStart.Format("20060102")}, "__")
}

func GenMessageLogTableNames(tx *gorm.DB, namespace snowflake.ID, startAt time.Time, endAt time.Time) []string {
	if startAt.After(endAt) {
		return nil
	}
	tableNames := make([]string, 0)
	firstMonday := getFirstMonday(startAt)
	for current := firstMonday; current.Before(endAt); current = current.AddDate(0, 0, 7) {
		if tableName := GenMessageLogTableName(namespace, current); HasTable(tx, tableName) {
			tableNames = append(tableNames, tableName)
		}
	}
	return tableNames
}

func HasTable(tx *gorm.DB, tableName string) bool {
	return tx.Migrator().HasTable(tableName)
}

func getFirstMonday(date time.Time) time.Time {
	offset := int(time.Monday - date.Weekday())
	if offset > 0 {
		offset -= 7
	}
	return date.AddDate(0, 0, offset)
}
