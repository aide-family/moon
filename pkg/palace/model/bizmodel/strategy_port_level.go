package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

// StrategyPort 策略端口
type StrategyPort struct {
	// 告警等级ID
	LevelID uint32 `json:"level_id,omitempty"`
	// 策略告警组
	AlarmNoticeGroupIds []uint32            `json:"alarm_group_ids,omitempty"`
	AlarmNoticeGroups   []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
	// 阈值
	Threshold int64 `json:"threshold,omitempty"`
	// 端口
	Port uint32 `json:"port,omitempty"`
	// 告警页面
	AlarmPageIds []uint32   `json:"alarm_page_ids,omitempty"`
	AlarmPage    []*SysDict `json:"alarm_page,omitempty"`
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
