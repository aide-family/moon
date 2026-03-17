package do

import (
	"strings"
	"time"

	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

const (
	TableNameAlertEvent = "alert_events"
)

// AlertEventStatus matches proto AlertEventStatus (1=firing, 2=intervened, 3=suppressed, 4=recovered).
const (
	AlertEventStatusUnknown    = 0
	AlertEventStatusFiring     = 1
	AlertEventStatusIntervened = 2
	AlertEventStatusSuppressed = 3
	AlertEventStatusRecovered  = 4
)

// Twitter/snowflake epoch (Nov 04 2010) in ms; used to derive time from ID.
const snowflakeEpochMs int64 = 1288834974657

type AlertEvent struct {
	EventBaseModel
	NamespaceUID        snowflake.ID                `gorm:"column:namespace_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	StrategyUID         snowflake.ID                `gorm:"column:strategy_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	StrategyGroupUID    snowflake.ID                `gorm:"column:strategy_group_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	LevelUID            snowflake.ID                `gorm:"column:level_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	Summary             string                      `gorm:"column:summary;type:varchar(500);default:''"`
	Description         string                      `gorm:"column:description;type:text;default:''"`
	Expr                string                      `gorm:"column:expr;type:text;default:''"`
	FiredAt             time.Time                   `gorm:"column:fired_at"`
	Value               float64                     `gorm:"column:value"`
	Labels              *safety.Map[string, string] `gorm:"column:labels;type:json;"`
	DatasourceUID       snowflake.ID                `gorm:"column:datasource_uid"`
	EvaluatorType       string                      `gorm:"column:evaluator_type;size:32;default:''"`
	EvaluatorSnapshotID snowflake.ID                `gorm:"column:evaluator_snapshot_id;index:idx__alert_event__namespace_uid__evaluator_snapshot_id"`
	Fingerprint         string                      `gorm:"column:fingerprint;size:64;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	Status              int32                       `gorm:"column:status;default:1"`
	IntervenedAt        *time.Time                  `gorm:"column:intervened_at"`
	IntervenedBy        snowflake.ID                `gorm:"column:intervened_by"`
	SuppressedUntil     *time.Time                  `gorm:"column:suppressed_until"`
	RecoveredAt         *time.Time                  `gorm:"column:recovered_at"`
	RecoveredBy         snowflake.ID                `gorm:"column:recovered_by"`
	EvaluatorSnapshot   *EvaluatorSnapshot          `gorm:"foreignKey:EvaluatorSnapshotID;references:ID"`
}

func (AlertEvent) TableName() string {
	return TableNameAlertEvent
}

// GenAlertEventTableName returns the shard table name: alert_events__{namespace}__{YYYYMMDD of Monday}.
func GenAlertEventTableName(namespace snowflake.ID, t time.Time) string {
	weekStart := getFirstMonday(t)
	return strings.Join([]string{TableNameAlertEvent, namespace.String(), weekStart.Format("20060102")}, "__")
}

// GenAlertEventTableNames returns existing table names for the namespace and time range (each week).
func GenAlertEventTableNames(tx *gorm.DB, namespace snowflake.ID, startAt, endAt time.Time) []string {
	if startAt.After(endAt) {
		return nil
	}
	names := make([]string, 0)
	firstMonday := getFirstMonday(startAt)
	for current := firstMonday; !current.After(endAt); current = current.AddDate(0, 0, 7) {
		tableName := GenAlertEventTableName(namespace, current)
		if tx.Migrator().HasTable(tableName) {
			names = append(names, tableName)
		}
	}
	return names
}

func getFirstMonday(date time.Time) time.Time {
	offset := int(time.Monday - date.Weekday())
	if offset > 0 {
		offset -= 7
	}
	return date.AddDate(0, 0, offset)
}

// AlertEventTimeFromID returns the time embedded in a snowflake ID (for shard lookup).
func AlertEventTimeFromID(id snowflake.ID) time.Time {
	ms := (id.Int64() >> 22) + snowflakeEpochMs
	return time.UnixMilli(ms)
}
