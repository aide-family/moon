package vobj

// UserMessageType 用户消息类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=UserMessageType -linecomment
type UserMessageType int

const (
	// UserMessageTypeUnknown 未知
	UserMessageTypeUnknown UserMessageType = iota // unknown

	// UserMessageTypeInfo 信息
	UserMessageTypeInfo // info

	// UserMessageTypeWarning 警告
	UserMessageTypeWarning // warning

	// UserMessageTypeError 错误
	UserMessageTypeError // error

	// UserMessageTypeSuccess 成功
	UserMessageTypeSuccess // success
)
