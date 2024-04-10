package vobj

import (
	"github.com/aide-family/moon/api"
)

type Category int32

const (
	// CategoryUnknown 未知
	CategoryUnknown Category = iota
	// CategoryPromLabel prometheus 标签
	CategoryPromLabel
	// CategoryPromAnnotation prometheus 描述信息
	CategoryPromAnnotation
	// CategoryPromStrategy prometheus 规则
	CategoryPromStrategy
	// CategoryPromStrategyGroup prometheus 规则组
	CategoryPromStrategyGroup
	// CategoryAlarmLevel 告警级别
	CategoryAlarmLevel
	// CategoryAlarmStatus 告警状态
	CategoryAlarmStatus
	// CategoryNotifyType 通知类型
	CategoryNotifyType
	//CategoryAlarmPage 告警页面
	CategoryAlarmPage
)

// String Category 转换为字符串
func (c Category) String() string {
	switch c {
	case CategoryPromLabel:
		return "prometheus 标签"
	case CategoryPromAnnotation:
		return "prometheus 描述信息"
	case CategoryPromStrategy:
		return "prometheus 规则"
	case CategoryPromStrategyGroup:
		return "prometheus 规则组"
	case CategoryAlarmLevel:
		return "告警等级"
	case CategoryAlarmStatus:
		return "告警状态"
	case CategoryNotifyType:
		return "通知类型"
	case CategoryAlarmPage:
		return "告警页面"
	case CategoryUnknown:
		fallthrough
	default:
		return "未知"
	}
}

// Key Category 转换为 key
func (c Category) Key() string {
	switch c {
	case CategoryPromLabel:
		return "prom_label"
	case CategoryPromAnnotation:
		return "prom_annotation"
	case CategoryPromStrategy:
		return "prom_strategy"
	case CategoryPromStrategyGroup:
		return "prom_strategy_group"
	case CategoryAlarmLevel:
		return "alarm_level"
	case CategoryAlarmStatus:
		return "alarm_status"
	case CategoryNotifyType:
		return "notify_type"
	case CategoryAlarmPage:
		return "alarm_page"
	default:
		return "unknown"
	}
}

// Value Category 转换为 value
func (c Category) Value() int32 {
	return int32(c)
}

// ApiCategory 转换为 api 枚举
func (c Category) ApiCategory() api.Category {
	return api.Category(c)
}

// IsUnknown 是否未知
func (c Category) IsUnknown() bool {
	return c == CategoryUnknown
}
