package vobj

// AlarmSendType 告警发送类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=AlarmSendType -linecomment
type AlarmSendType int

const (
	// AlarmSendTypeUnknown 未知
	AlarmSendTypeUnknown AlarmSendType = iota // 未知
	// AlarmSendTypeEmail 邮件
	AlarmSendTypeEmail // 邮件
	// AlarmSendTypeSMS 短信
	AlarmSendTypeSMS // 短信
	// AlarmSendTypeDingTalk 钉钉
	AlarmSendTypeDingTalk // 钉钉
	// AlarmSendTypeFeiShu 飞书
	AlarmSendTypeFeiShu // 飞书
	// AlarmSendTypeWechat 微信
	AlarmSendTypeWechat // 微信
	// AlarmSendTypeCustom 自定义
	AlarmSendTypeCustom // 自定义
)
