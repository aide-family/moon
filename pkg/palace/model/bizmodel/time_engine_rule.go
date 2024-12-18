package bizmodel

import (
	"strconv"
	"strings"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameTimeEngineRule = "time_engine_rule"

// TimeEngineRule 时间引擎规则
type TimeEngineRule struct {
	model.AllFieldModel
	Name     string                  `gorm:"column:name;type:varchar(64);not null;comment:规则名称" json:"name"`    // 规则名称
	Remark   string                  `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"` // 备注
	Status   vobj.Status             `gorm:"column:status;type:int;not null;comment:状态" json:"status"`          // 状态
	Category vobj.TimeEngineRuleType `gorm:"column:category;type:int;not null;comment:规则类型" json:"category"`    // 规则类型
	Rule     string                  `gorm:"column:rule;type:text;not null;comment:规则,分割数字" json:"rule"`        // 规则
}

// Matcher 匹配器
func (c *TimeEngineRule) Matcher() types.Matcher {
	rule := c.getRule()
	if len(rule) == 0 {
		return nil
	}
	switch c.Category {
	case vobj.TimeEngineRuleTypeHourRange:
		if len(rule) < 2 {
			return nil
		}
		return &types.HourRange{
			Start: rule[0],
			End:   rule[1],
		}
	case vobj.TimeEngineRuleTypeDaysOfWeek:
		daysOfWeek := types.DaysOfWeek(rule)
		return &daysOfWeek
	case vobj.TimeEngineRuleTypeDaysOfMonth:
		if len(rule) < 2 {
			return nil
		}
		return &types.DaysOfMonth{
			Start: rule[0],
			End:   rule[1],
		}
	case vobj.TimeEngineRuleTypeMonths:
		if len(rule) < 2 {
			return nil
		}
		return &types.Months{
			Start: rule[0],
			End:   rule[1],
		}
	default:
		return nil
	}
}

// String json string
func (c *TimeEngineRule) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *TimeEngineRule) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *TimeEngineRule) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamRole's table name
func (*TimeEngineRule) TableName() string {
	return tableNameTimeEngineRule
}

// getRule 获取规则
func (c *TimeEngineRule) getRule() []int {
	if c.Rule == "" {
		return nil
	}
	rule := strings.Split(c.Rule, ",")
	nums := make([]int, 0, len(rule))
	for _, r := range rule {
		num, err := strconv.Atoi(r)
		if err != nil {
			continue
		}
		nums = append(nums, num)
	}
	return nums
}
