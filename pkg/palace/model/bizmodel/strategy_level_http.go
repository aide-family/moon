package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyHTTPLevel HTTP监控策略定义， 用于监控指定URL的响应时间、状态码
type StrategyHTTPLevel struct {
	// 状态码
	StatusCode string `json:"statusCode,omitempty"`
	// 响应时间
	ResponseTime float64 `json:"responseTime,omitempty"`
	// 请求头
	Headers []*vobj.Header `json:"headers,omitempty"`
	// 请求body
	Body string `json:"body,omitempty"`
	// 请求方式
	Method vobj.HTTPMethod `json:"method,omitempty"`
	// 查询参数
	QueryParams string `json:"queryParams,omitempty"`
	// 状态码判断条件
	StatusCodeCondition vobj.Condition `json:"condition,omitempty"`
	// 响应时间判断条件
	ResponseTimeCondition vobj.Condition `json:"responseTimeCondition,omitempty"`

	// 告警等级ID
	Level *SysDict `json:"level,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
}

// String 字符串
func (s *StrategyHTTPLevel) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyHTTPLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyHTTPLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
