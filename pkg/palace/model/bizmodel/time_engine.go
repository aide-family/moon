package bizmodel

import (
	"time"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameTimeEngine = "time_engine"

var _ types.TimeEngineer = (*TimeEngine)(nil)

// TimeEngine 时间引擎
type TimeEngine struct {
	model.AllFieldModel
	Name   string            `gorm:"column:name;type:varchar(64);not null;comment:规则名称" json:"name"`    // 规则名称
	Remark string            `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"` // 备注
	Status vobj.Status       `gorm:"column:status;type:int;not null;comment:状态" json:"status"`          // 状态
	Rules  []*TimeEngineRule `gorm:"many2many:time_engine_rule_relation;" json:"rules"`
}

// IsAllowed 判断条件是否允许
func (c *TimeEngine) IsAllowed(time time.Time) bool {
	if c == nil || len(c.Rules) == 0 {
		return true
	}

	return c.matches(time)
}

func (c *TimeEngine) matches(t time.Time) bool {
	configs := types.SliceTo(c.Rules, func(r *TimeEngineRule) types.Matcher { return r.Matcher() })
	return types.NewTimeEngine(types.WithConfigurations(configs)).IsAllowed(t)
}

// String json string
func (c *TimeEngine) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *TimeEngine) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *TimeEngine) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamRole's table name
func (*TimeEngine) TableName() string {
	return tableNameTimeEngine
}
