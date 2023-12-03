package valueobj

import (
	"database/sql/driver"
	"encoding/json"

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
)

const (
	// NotifyTypeUnknown 未知
	NotifyTypeUnknown NotifyType = iota
	// NotifyTypeEmail 邮件
	NotifyTypeEmail
	// NotifyTypeSms 短信
	NotifyTypeSms
	// NotifyTypePhone 电话
	NotifyTypePhone
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

// String 转换为字符串
func (a NotifyType) String() string {
	switch a {
	case NotifyTypeEmail:
		return "邮件"
	case NotifyTypeSms:
		return "短信"
	case NotifyTypePhone:
		return "电话"
	case NotifyTypeUnknown:
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

// Key 转换为键
func (a NotifyType) Key() string {
	switch a {
	case NotifyTypeEmail:
		return "email"
	case NotifyTypeSms:
		return "sms"
	case NotifyTypePhone:
		return "phone"
	case NotifyTypeUnknown:
		return "unknown"
	default:
		return "unknown"
	}
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

type NotifyTypes []NotifyType

func (l *NotifyTypes) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}

	str, err := json.Marshal(l)
	return string(str), err
}

func (l *NotifyTypes) Scan(src any) error {
	return json.Unmarshal(src.([]byte), l)
}
