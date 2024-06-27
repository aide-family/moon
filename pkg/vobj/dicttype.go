package vobj

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DictType int32

const (
	// DictTypeUnknown 未知
	DictTypeUnknown DictType = iota
	// DictTypePromLabel prometheus 标签
	DictTypePromLabel
	// DictTypePromAnnotation prometheus 描述信息
	DictTypePromAnnotation
	// DictTypePromStrategy prometheus 规则
	DictTypePromStrategy
	// DictTypePromStrategyGroup prometheus 规则组
	DictTypePromStrategyGroup
	// DictTypeAlarmLevel 告警级别
	DictTypeAlarmLevel
	// DictTypeAlarmStatus 告警状态
	DictTypeAlarmStatus
	// DictTypeNotifyType 通知类型
	DictTypeNotifyType
	// DictTypeAlarmPage 告警页面
	DictTypeAlarmPage
)

// Enum value maps for DictType.
var (
	DictType_name = map[int32]string{
		0: "DICTTYPE_UNKNOWN",
		1: "DICTTYPE_PROM_LABEL",
		2: "DICTTYPE_PROM_ANNOTATION",
		3: "DICTTYPE_PROM_STRATEGY",
		4: "DICTTYPE_PROM_STRATEGY_GROUP",
		5: "DICTTYPE_ALARM_LEVEL",
		6: "DICTTYPE_ALARM_STATUS",
		7: "DICTTYPE_NOTIFY_TYPE",
	}
	DictType_value = map[string]int32{
		"DICTTYPE_UNKNOWN":             0,
		"DICTTYPE_PROM_LABEL":          1,
		"DICTTYPE_PROM_ANNOTATION":     2,
		"DICTTYPE_PROM_STRATEGY":       3,
		"DICTTYPE_PROM_STRATEGY_GROUP": 4,
		"DICTTYPE_ALARM_LEVEL":         5,
		"DICTTYPE_ALARM_STATUS":        6,
		"DICTTYPE_NOTIFY_TYPE":         7,
	}
)

func (x DictType) Enum() *DictType {
	p := new(DictType)
	*p = x
	return p
}

func (x DictType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// String DictType 转换为字符串
func (d DictType) String() string {
	switch d {
	case DictTypePromLabel:
		return "prometheus 标签"
	case DictTypePromAnnotation:
		return "prometheus 描述信息"
	case DictTypePromStrategy:
		return "prometheus 规则"
	case DictTypePromStrategyGroup:
		return "prometheus 规则组"
	case DictTypeAlarmLevel:
		return "告警等级"
	case DictTypeAlarmStatus:
		return "告警状态"
	case DictTypeNotifyType:
		return "通知类型"
	case DictTypeAlarmPage:
		return "告警页面"
	case DictTypeUnknown:
		fallthrough
	default:
		return "未知"
	}
}

// Key DictType 转换为 key
func (d DictType) Key() string {
	switch d {
	case DictTypePromLabel:
		return "prom_label"
	case DictTypePromAnnotation:
		return "prom_annotation"
	case DictTypePromStrategy:
		return "prom_strategy"
	case DictTypePromStrategyGroup:
		return "prom_strategy_group"
	case DictTypeAlarmLevel:
		return "alarm_level"
	case DictTypeAlarmStatus:
		return "alarm_status"
	case DictTypeNotifyType:
		return "notify_type"
	case DictTypeAlarmPage:
		return "alarm_page"
	default:
		return "unknown"
	}
}

// Value DictType 转换为 value
func (c DictType) Value() int32 {
	return int32(c)
}

// IsUnknown 是否未知
func (c DictType) IsUnknown() bool {
	return c == DictTypeUnknown
}

func (i DictType) GetValue() int {
	return int(i)
}
