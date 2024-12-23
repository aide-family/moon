package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyDomainLevel 域名证书｜端口 等级策略明细
type StrategyDomainLevel struct {
	// 告警等级I
	Level *SysDict `json:"level,omitempty"`
	// 阈值 （证书类型就是剩余天数，端口就是0：关闭，1：开启）
	Threshold int64 `json:"threshold,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarmGroupList,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
	// 判断条件
	Condition vobj.Condition `json:"condition,omitempty"`
}

// String 字符串
func (s *StrategyDomainLevel) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyDomainLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyDomainLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
