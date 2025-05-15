package team

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timer"
)

var _ do.TimeEngineRule = (*TimeEngineRule)(nil)

const tableNameTimeEngineRule = "team_time_engine_rules"

type TimeEngineRule struct {
	do.TeamModel
	Name    string                  `gorm:"column:name;type:varchar(64);not null;comment:名称" json:"name"`
	Remark  string                  `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status  vobj.GlobalStatus       `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Rule    Rules                   `gorm:"column:rule;type:text;not null;comment:规则" json:"rule"`
	Type    vobj.TimeEngineRuleType `gorm:"column:type;type:tinyint(2);not null;comment:类型" json:"type"`
	Engines []*TimeEngine           `gorm:"many2many:team_time_engine__time_rules" json:"engines"`
}

func (t *TimeEngineRule) GetName() string {
	if t == nil {
		return ""
	}
	return t.Name
}

func (t *TimeEngineRule) GetRemark() string {
	if t == nil {
		return ""
	}
	return t.Remark
}

func (t *TimeEngineRule) GetStatus() vobj.GlobalStatus {
	if t == nil {
		return vobj.GlobalStatusUnknown
	}
	return t.Status
}

func (t *TimeEngineRule) GetTimeEngines() []do.TimeEngine {
	if t == nil {
		return nil
	}
	return slices.Map(t.Engines, func(e *TimeEngine) do.TimeEngine { return e })
}

func (t *TimeEngineRule) GetType() vobj.TimeEngineRuleType {
	if t == nil {
		return vobj.TimeEngineRuleTypeUnknown
	}
	return t.Type
}

func (t *TimeEngineRule) GetRules() []int {
	if t == nil {
		return nil
	}
	return t.Rule
}

func (t *TimeEngineRule) TableName() string {
	return tableNameTimeEngineRule
}

type Rules []int

func (r Rules) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *Rules) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), r)
}

func (t *TimeEngineRule) ToTimerMatcher() (timer.Matcher, error) {
	switch t.Type {
	case vobj.TimeEngineRuleTypeHourRange:
		return timer.NewHourRange(t.Rule)
	case vobj.TimeEngineRuleTypeDaysOfWeek:
		return timer.NewDaysOfWeek(t.Rule)
	case vobj.TimeEngineRuleTypeDayOfMonth:
		return timer.NewDayOfMonths(t.Rule)
	case vobj.TimeEngineRuleTypeMonth:
		return timer.NewMonth(t.Rule)
	case vobj.TimeEngineRuleTypeHourMinuteRange:
		return timer.NewHourMinuteRangeWithSlice(t.Rule)
	case vobj.TimeEngineRuleTypeHour:
		return timer.NewHour(t.Rule)
	default:
		return nil, merr.ErrorParamsError("invalid time engine rule type: %d", t.Type)
	}
}
