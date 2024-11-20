package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

// 证书过期、端口开闭监控策略
const tableNameStrategyDomain = "strategy_domain"

// StrategyDomain 域名证书｜端口 等级策略明细
type StrategyDomain struct {
	model.AllFieldModel
	// 所属策略
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__domain__strategy_id__level_id" json:"strategy_id"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID" json:"strategy"`
	// 告警等级ID
	LevelID uint32   `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级ID" json:"level_id"`
	Level   *SysDict `gorm:"foreignKey:LevelID" json:"level"`
	// 执行频率
	Interval uint32 `gorm:"column:interval;type:int unsigned;not null;comment:执行频率seconds" json:"interval"`
	// 阈值 （证书类型就是剩余天数，端口就是0：关闭，1：开启）
	Threshold uint32 `gorm:"column:threshold;type:int unsigned;not null;comment:阈值" json:"threshold"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_domain_alarm_groups;" json:"alarm_groups"`
}

// TableName 表名
func (*StrategyDomain) TableName() string {
	return tableNameStrategyDomain
}

// String 字符串
func (s *StrategyDomain) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyDomain) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyDomain) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
