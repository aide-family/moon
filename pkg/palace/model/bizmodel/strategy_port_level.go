package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameStrategyPort = "strategy_port_level"

type StrategyPort struct {
	model.AllFieldModel

	// 告警等级ID
	LevelID uint32   `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级ID" json:"level_id"`
	Level   *SysDict `gorm:"foreignKey:LevelID" json:"level"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_port_alarm_groups;" json:"alarm_groups"`
	// 阈值
	Threshold int64 `gorm:"column:threshold;type:bigint;not null;comment:阈值" json:"threshold"`
	// 端口
	Port uint32 `gorm:"column:port;type:int unsigned;not null;comment:端口" json:"port"`
	// 告警页面
	AlarmPage []*SysDict `gorm:"many2many:strategy_port_level_alarm_pages" json:"alarm_page"`
}

// TableName 表名
func (*StrategyPort) TableName() string {
	return tableNameStrategyPort
}

// String 字符串
func (s *StrategyPort) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyPort) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyPort) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
