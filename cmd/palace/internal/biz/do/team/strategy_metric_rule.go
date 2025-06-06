package team

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.StrategyMetricRule = (*StrategyMetricRule)(nil)

const tableNameStrategyMetricRule = "team_strategy_metric_rules"

type StrategyMetricRule struct {
	do.TeamModel
	StrategyID       uint32                           `gorm:"column:strategy_id;type:int unsigned;not null;comment:strategy ID" json:"strategyID"`
	StrategyMetricID uint32                           `gorm:"column:strategy_metric_id;type:int unsigned;not null;comment:strategy metric ID" json:"strategyMetricID"`
	StrategyMetric   *StrategyMetric                  `gorm:"foreignKey:StrategyMetricID;references:ID" json:"strategyMetric"`
	LevelID          uint32                           `gorm:"column:level_id;type:int unsigned;not null;comment:level ID" json:"levelID"`
	Level            *Dict                            `gorm:"foreignKey:LevelID;references:ID" json:"level"`
	SampleMode       vobj.SampleMode                  `gorm:"column:sample_mode;type:tinyint(2);not null;comment:sample mode" json:"sampleMode"`
	Condition        vobj.ConditionMetric             `gorm:"column:condition;type:tinyint(2);not null;comment:condition" json:"condition"`
	Total            int64                            `gorm:"column:total;type:bigint;not null;comment:sample count" json:"total"`
	Values           Values                           `gorm:"column:values;type:json;not null;comment:values" json:"values"`
	Duration         int64                            `gorm:"column:duration;type:bigint;not null;comment:duration" json:"duration"`
	Status           vobj.GlobalStatus                `gorm:"column:status;type:tinyint(2);not null;comment:status" json:"status"`
	Notices          []*NoticeGroup                   `gorm:"many2many:team_strategy_metric_rule_notice_groups" json:"notices"`
	LabelNotices     []*StrategyMetricRuleLabelNotice `gorm:"foreignKey:StrategyMetricRuleID;references:ID" json:"labelNotices"`
	AlarmPages       []*Dict                          `gorm:"many2many:team_strategy_metric_rule_alarm_pages" json:"alarmPages"`
}

func (r *StrategyMetricRule) GetStrategyID() uint32 {
	if r == nil {
		return 0
	}
	return r.StrategyID
}

func (r *StrategyMetricRule) GetStrategyMetricID() uint32 {
	if r == nil {
		return 0
	}
	return r.StrategyMetricID
}

func (r *StrategyMetricRule) GetStrategyMetric() do.StrategyMetric {
	if r == nil {
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
	if r == nil {
		return nil
	}
	return r.Values
}

func (r *StrategyMetricRule) GetDuration() int64 {
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
