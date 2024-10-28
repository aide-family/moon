package vobj

import (
	"strings"
)

// NotifyType 通知类型
type NotifyType int

const (
	// NotifyTypePhone 电话通知
	NotifyTypePhone NotifyType = 1 << iota

	// NotifyTypeSMS 短信通知
	NotifyTypeSMS

	// NotifyTypeEmail 邮件通知
	NotifyTypeEmail
)

// IsPhone 是否电话通知
func (n NotifyType) IsPhone() bool {
	return n&NotifyTypePhone != 0
}

// IsSMS 是否短信通知
func (n NotifyType) IsSMS() bool {
	return n&NotifyTypeSMS != 0
}

// IsEmail 是否邮件通知
func (n NotifyType) IsEmail() bool {
	return n&NotifyTypeEmail != 0
}

func (n NotifyType) String() string {
	notify := make([]string, 0, 4)
	if n.IsPhone() {
		notify = append(notify, "电话")
	}
	if n.IsSMS() {
		notify = append(notify, "短信")
	}
	if n.IsEmail() {
		notify = append(notify, "邮件")
	}
	if len(notify) == 0 {
		return "未知"
	}
	return strings.Join(notify, ",")
}

func (n NotifyType) EnString() string {
	notify := make([]string, 0, 4)
	if n.IsPhone() {
		notify = append(notify, "phone")
	}
	if n.IsSMS() {
		notify = append(notify, "sms")
	}
	if n.IsEmail() {
		notify = append(notify, "email")
	}
	if len(notify) == 0 {
		return "unknown"
	}
	return strings.Join(notify, ",")
}

// GetValue 获取值
func (n NotifyType) GetValue() int {
	return int(n)
}
