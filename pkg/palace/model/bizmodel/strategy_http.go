package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyHTTP HTTP监控策略定义， 用于监控指定URL的响应时间、状态码
type StrategyHTTP struct {
	// 所属策略
	StrategyID uint32 `json:"strategy_id,omitempty"`
	// 告警等级ID
	LevelID uint32   `json:"level_id,omitempty"`
	Level   *SysDict `json:"level,omitempty"`
	// 策略告警组
	NoticeGroupIds []uint32 `json:"noticeGroupIds,omitempty"`

	AlarmNoticeGroups []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
	// 状态码
	StatusCodes uint32 `json:"status_codes,omitempty"`
	// 响应时间
	ResponseTime uint32 `json:"response_time,omitempty"`
	// 请求头
	Headers []*vobj.Header `json:"headers,omitempty"`
	// 请求body
	Body string `json:"body,omitempty"`
	// 请求方式
	Method vobj.HTTPMethod `json:"method,omitempty"`
	// 查询参数
	QueryParams string `json:"query_params,omitempty"`
	// 状态码判断条件
	StatusCodeCondition vobj.Condition `json:"condition,omitempty"`
	// 响应时间判断条件
	ResponseTimeCondition vobj.Condition `json:"response_time_condition,omitempty"`
}

// String 字符串
func (s *StrategyHTTP) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyHTTP) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyHTTP) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
