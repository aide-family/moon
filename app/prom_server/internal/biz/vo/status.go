package vo

import (
	"prometheus-manager/api"
)

type Status int32

const (
	// StatusUnknown 未知状态
	StatusUnknown Status = iota
	// StatusEnabled 启用
	StatusEnabled
	// StatusDisabled 禁用
	StatusDisabled
)

// IsUnknown 是否未知状态
func (s Status) IsUnknown() bool {
	return s == StatusUnknown
}

// String 获取状态字符串
func (s Status) String() string {
	switch s {
	case StatusEnabled:
		return "启用"
	case StatusDisabled:
		return "禁用"
	default:
		return "未知"
	}
}

// Key 获取状态key
func (s Status) Key() string {
	switch s {
	case StatusEnabled:
		return "enabled"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// Value 获取状态值
func (s Status) Value() int32 {
	return int32(s)
}

// ApiStatus 转换为api状态
func (s Status) ApiStatus() api.Status {
	return api.Status(s)
}

type IsSendRecover uint8

const (
	// IsSendRecoverUnknown 未知
	IsSendRecoverUnknown IsSendRecover = iota
	// IsSendRecoverYes 发送恢复通知
	IsSendRecoverYes
	// IsSendRecoverNo 不发送恢复通知
	IsSendRecoverNo
)

// String 获取状态字符串
func (s IsSendRecover) String() string {
	switch s {
	case IsSendRecoverUnknown:
		return "未知"
	case IsSendRecoverYes:
		return "发送恢复通知"
	case IsSendRecoverNo:
		return "不发送恢复通知"
	default:
		return "未定义"
	}
}

// Value 获取状态值
func (s IsSendRecover) Value() bool {
	return s == IsSendRecoverYes
}

// NewIsSendRecover 转换为IsSendRecover
func NewIsSendRecover(value bool) IsSendRecover {
	if value {
		return IsSendRecoverYes
	}
	return IsSendRecoverNo
}
