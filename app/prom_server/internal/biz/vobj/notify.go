package vobj

import (
	"prometheus-manager/api"
)

type NotifyApp int32
type NotifyType int32

const (
	// NotifyAppUnknown 未知
	NotifyAppUnknown NotifyApp = iota
	// NotifyAppDingDing 钉钉
	NotifyAppDingDing
	// NotifyAppWeChatWork 企业微信
	NotifyAppWeChatWork
	// NotifyAppFeiShu 飞书
	NotifyAppFeiShu
	// NotifyAppCustom 自定义
	NotifyAppCustom
)

const (
	// NotifyTypeUnknown 未知
	NotifyTypeUnknown NotifyType = iota
	// NotifyTypeEmail 邮件
	NotifyTypeEmail NotifyType = 1 << iota
	// NotifyTypeSms 短信
	NotifyTypeSms NotifyType = 1 << iota
	// NotifyTypePhone 电话
	NotifyTypePhone NotifyType = 1 << iota
)

// String 转换为字符串
func (a NotifyApp) String() string {
	switch a {
	case NotifyAppDingDing:
		return "钉钉"
	case NotifyAppWeChatWork:
		return "企业微信"
	case NotifyAppFeiShu:
		return "飞书"
	case NotifyAppCustom:
		return "自定义"
	case NotifyAppUnknown:
		return "未知"
	default:
		return "未知"
	}
}

// Key 转换为键
func (a NotifyApp) Key() string {
	switch a {
	case NotifyAppDingDing:
		return "dingding"
	case NotifyAppWeChatWork:
		return "wechatwork"
	case NotifyAppFeiShu:
		return "feishu"
	case NotifyAppCustom:
		return "custom"
	case NotifyAppUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// IsUnknown 是否未知
func (a NotifyApp) IsUnknown() bool {
	return a == NotifyAppUnknown
}

// Value 转换为值
func (a NotifyApp) Value() int32 {
	return int32(a)
}

// Value 转换为值
func (a NotifyType) Value() int32 {
	return int32(a)
}

// ApiNotifyApp API通知应用
func (a NotifyApp) ApiNotifyApp() api.NotifyApp {
	return api.NotifyApp(a)
}

// ApiNotifyType API通知类型
func (a NotifyType) ApiNotifyType() api.NotifyType {
	return api.NotifyType(a)
}

// IsEmail 邮件
func (a NotifyType) IsEmail() bool {
	return a&NotifyTypeEmail == NotifyTypeEmail
}

// IsSms 短信
func (a NotifyType) IsSms() bool {
	return a&NotifyTypeSms == NotifyTypeSms
}

// IsPhone 电话
func (a NotifyType) IsPhone() bool {
	return a&NotifyTypePhone == NotifyTypePhone
}
