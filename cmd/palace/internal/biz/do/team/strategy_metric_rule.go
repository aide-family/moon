package team

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

var _ do.StrategyMetricRule = (*StrategyMetricRule)(nil)

const tableNameStrategyMetricRule = "team_strategy_metric_rules"

type StrategyMetricRule struct {
	do.TeamModel
	StrategyMetricID uint32                           `gorm:"column:strategy_metric_id;type:int unsigned;not null;comment:策略指标id" json:"strategyMetricID"`
	StrategyMetric   *StrategyMetric                  `gorm:"foreignKey:StrategyMetricID;references:ID" json:"strategyMetric"`
	LevelID          uint32                           `gorm:"column:level_id;type:int unsigned;not null;comment:等级id" json:"levelID"`
	Level            *Dict                            `gorm:"foreignKey:LevelID;references:ID" json:"level"`
	SampleMode       vobj.SampleMode                  `gorm:"column:sample_mode;type:tinyint(2);not null;comment:采样方式" json:"sampleMode"`
	Condition        vobj.ConditionMetric             `gorm:"column:condition;type:tinyint(2);not null;comment:条件" json:"condition"`
	Total            int64                            `gorm:"column:total;type:bigint;not null;comment:采样数量" json:"total"`
	Values           Values                           `gorm:"column:values;type:json;not null;comment:值" json:"values"`
	Duration         time.Duration                    `gorm:"column:duration;type:bigint(20);not null;comment:持续时间" json:"duration"`
	Status           vobj.GlobalStatus                `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Notices          []*NoticeGroup                   `gorm:"many2many:team_strategy_metric_rule_notice_groups" json:"notices"`
	LabelNotices     []*StrategyMetricRuleLabelNotice `gorm:"foreignKey:StrategyMetricRuleID;references:ID" json:"labelNotices"`
	AlarmPages       []*Dict                          `gorm:"many2many:team_strategy_metric_rule_alarm_pages" json:"alarmPages"`
}

func (r *StrategyMetricRule) GetStrategyMetricID() uint32 {
	if r == nil {
		return 0
	}
	return r.StrategyMetricID
}

func (r *StrategyMetricRule) GetStrategyMetric() do.StrategyMetric {
	if r == nil || r.StrategyMetric == nil {
		return nil
	}
	return r.StrategyMetric
}

func (r *StrategyMetricRule) GetLevelID() uint32 {
	if r == nil {
		return 0
	}
	return r.LevelID
}

func (r *StrategyMetricRule) GetLevel() do.TeamDict {
	if r == nil {
		return nil
	}
	return r.Level
}

func (r *StrategyMetricRule) GetSampleMode() vobj.SampleMode {
	if r == nil {
		return vobj.SampleModeUnknown
	}
	return r.SampleMode
}

func (r *StrategyMetricRule) GetCondition() vobj.ConditionMetric {
	if r == nil {
		return vobj.ConditionMetricUnknown
	}
	return r.Condition
}

func (r *StrategyMetricRule) GetTotal() int64 {
	if r == nil {
		return 0
	}
	return r.Total
}

func (r *StrategyMetricRule) GetValues() []float64 {
	if r == nil || r.Values == nil {
		return nil
	}
	return r.Values
}

func (r *StrategyMetricRule) GetDuration() time.Duration {
	if r == nil {
		return 0
	}
	return r.Duration
}

func (r *StrategyMetricRule) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	return r.Status
}

func (r *StrategyMetricRule) GetNotices() []do.NoticeGroup {
	if r == nil {
		return nil
	}
	return slices.Map(r.Notices, func(v *NoticeGroup) do.NoticeGroup { return v })
}

func (r *StrategyMetricRule) GetLabelNotices() []do.StrategyMetricRuleLabelNotice {
	if r == nil {
		return nil
	}
	return slices.Map(r.LabelNotices, func(v *StrategyMetricRuleLabelNotice) do.StrategyMetricRuleLabelNotice { return v })
}

func (r *StrategyMetricRule) GetAlarmPages() []do.TeamDict {
	if r == nil {
		return nil
	}
	return slices.Map(r.AlarmPages, func(v *Dict) do.TeamDict { return v })
}

func (r *StrategyMetricRule) TableName() string {
	return tableNameStrategyMetricRule
}

var _ do.ORMModel = (*Values)(nil)

type Values []float64

func (v *Values) Scan(src any) error {
	switch value := src.(type) {
	case []byte:
		return json.Unmarshal(value, v)
	case string:
		return json.Unmarshal([]byte(value), v)
	default:
		return nil
	}
}

func (v Values) Value() (driver.Value, error) {
	return json.Marshal(v)
}
