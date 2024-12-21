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
	// 告警等级ID
	LevelID uint32   `json:"level_id,omitempty"`
	Level   *SysDict `json:"level"`
	// 阈值 （证书类型就是剩余天数，端口就是0：关闭，1：开启）
	Threshold           int64   `json:"threshold,omitempty"`
	AlarmNoticeGroupIds []int64 `json:"alarm_notice_group_ids,omitempty"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_domain_alarm_groups;" json:"alarm_groups"`
	// 告警页面ID
	AlarmPageIds []int64 `json:"alarm_page_ids,omitempty"`
	// 告警页面
	AlarmPage []*SysDict `gorm:"many2many:strategy_domain_level_alarm_pages" json:"alarm_page"`
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
