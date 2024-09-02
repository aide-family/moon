package vobj

import (
	"strings"
)

// NotifyType 通知类型
type NotifyType int

const (
	// Phone 电话通知
	Phone = 1 << iota

	// SMS 短信通知
	SMS

	// Email 邮件通知
	Email
)

// IsPhone 是否电话通知
func (n NotifyType) IsPhone() bool {
	return n&Phone != 0
}

// IsSMS 是否短信通知
func (n NotifyType) IsSMS() bool {
	return n&SMS != 0
}

// IsEmail 是否邮件通知
func (n NotifyType) IsEmail() bool {
	return n&Email != 0
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

func (n NotifyType) GetValue() int {
	return int(n)
}
