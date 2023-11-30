package valueobj

import (
	"prometheus-manager/api"
)

type NotifyApp int32

const (
	// NotifyAppUnknown 未知
	NotifyAppUnknown NotifyApp = iota
	// NotifyAppDingDing 钉钉
	NotifyAppDingDing
	// NotifyAppWeChatWork 企业微信
	NotifyAppWeChatWork
	// NotifyAppFeiShu 飞书
	NotifyAppFeiShu
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
	case NotifyAppUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// Value 转换为值
func (a NotifyApp) Value() int32 {
	return int32(a)
}

// ApiNotifyApp API通知应用
func (a NotifyApp) ApiNotifyApp() api.NotifyApp {
	return api.NotifyApp(a)
}
