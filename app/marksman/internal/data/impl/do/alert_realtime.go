package do

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
)

const (
	TableNameAlertRealtime = "alert_realtime"
)

type AlertRealtime struct {
	EventBaseModel
	NamespaceUID        snowflake.ID                `gorm:"column:namespace_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	StrategyGroupUID    snowflake.ID                `gorm:"column:strategy_group_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	StrategyGroupName   string                      `gorm:"column:strategy_group_name;type:varchar(100);default:''"`
	StrategyUID         snowflake.ID                `gorm:"column:strategy_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	StrategyName        string                      `gorm:"column:strategy_name;type:varchar(100);default:''"`
	LevelUID            snowflake.ID                `gorm:"column:level_uid;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	LevelName           string                      `gorm:"column:level_name;type:varchar(100);default:''"`
	BgColor             string                      `gorm:"column:bg_color;type:varchar(100);default:''"`
	DatasourceUID       snowflake.ID                `gorm:"column:datasource_uid"`
	DatasourceName      string                      `gorm:"column:datasource_name;type:varchar(100);default:''"`
	DatasourceLevelName string                      `gorm:"column:datasource_level_name;type:varchar(100);default:''"`
	Summary             string                      `gorm:"column:summary;type:varchar(500);default:''"`
	Description         string                      `gorm:"column:description;type:text;default:''"`
	Expr                string                      `gorm:"column:expr;type:text;default:''"`
	FiredAt             time.Time                   `gorm:"column:fired_at"`
	Value               float64                     `gorm:"column:value"`
	Labels              *safety.Map[string, string] `gorm:"column:labels;type:json;"`
	EvaluatorType       string                      `gorm:"column:evaluator_type;size:32;default:''"`
	EvaluatorSnapshotID snowflake.ID                `gorm:"column:evaluator_snapshot_id;index:idx__alert_event__namespace_uid__evaluator_snapshot_id"`
	Fingerprint         string                      `gorm:"column:fingerprint;size:64;index:idx__alert_event__namespace_uid__strategy_group_uid__strategy_uid__level_uid__fingerprint"`
	Status              enum.AlertEventStatus       `gorm:"column:status;default:1"`
	IntervenedAt        *time.Time                  `gorm:"column:intervened_at"`
	IntervenedBy        snowflake.ID                `gorm:"column:intervened_by"`
	IntervenedByName    string                      `gorm:"column:intervened_by_name;type:varchar(100);default:''"`
	SuppressedUntilAt   *time.Time                  `gorm:"column:suppressed_until"`
	SuppressedBy        snowflake.ID                `gorm:"column:suppressed_by"`
	SuppressedByName    string                      `gorm:"column:suppressed_by_name;type:varchar(100);default:''"`
	SuppressedReason    string                      `gorm:"column:suppressed_reason;type:text;default:''"`
	RecoveredAt         *time.Time                  `gorm:"column:recovered_at"`
	RecoveredBy         snowflake.ID                `gorm:"column:recovered_by"`
	RecoveredByName     string                      `gorm:"column:recovered_by_name;type:varchar(100);default:''"`
	RecoveredReason     string                      `gorm:"column:recovered_reason;type:text;default:''"`
	EvaluatorSnapshot   *EvaluatorSnapshot          `gorm:"foreignKey:EvaluatorSnapshotID;references:ID"`
}

func (AlertRealtime) TableName() string {
	return TableNameAlertRealtime
}
