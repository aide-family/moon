package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

// Ping监控策略定义， 用于监控指定IP的网络延迟、丢包率等
const tableNameStrategyPing = "strategy_ping"

type StrategyPing struct {
	model.AllFieldModel
	// 所属策略
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__ping__strategy_id__level_id" json:"strategy_id"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID" json:"strategy"`
	// 告警等级ID
	LevelID uint32   `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级ID" json:"level_id"`
	Level   *SysDict `gorm:"foreignKey:LevelID" json:"level"`
	// 执行频率
	Interval uint32 `gorm:"column:interval;type:int unsigned;not null;comment:执行频率seconds" json:"interval"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_ping_alarm_groups;" json:"alarm_groups"`
	// 总包数
	Total uint32 `gorm:"column:total;type:int unsigned;not null;comment:总包数" json:"total"`
	// 成功包数
	Success uint32 `gorm:"column:success;type:int unsigned;not null;comment:成功包数" json:"success"`
	// 丢包率
	LossRate float64 `gorm:"column:loss_rate;type:float;not null;comment:丢包率" json:"loss_rate"`
	// 平均延迟
	AvgDelay uint32 `gorm:"column:avg_delay;type:int unsigned;not null;comment:平均延迟" json:"avg_delay"`
	// 最大延迟
	MaxDelay uint32 `gorm:"column:max_delay;type:int unsigned;not null;comment:最大延迟" json:"max_delay"`
	// 最小延迟
	MinDelay uint32 `gorm:"column:min_delay;type:int unsigned;not null;comment:最小延迟" json:"min_delay"`
	// 标准差
	StdDev uint32 `gorm:"column:std_dev;type:int unsigned;not null;comment:标准差" json:"std_dev"`
}

// TableName 表名
func (*StrategyPing) TableName() string {
	return tableNameStrategyPing
}

// String 字符串
func (s *StrategyPing) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyPing) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyPing) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
