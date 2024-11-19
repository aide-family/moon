package vobj

// SendStatus 发送状态
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=SendStatus -linecomment
type SendStatus int

const (
	// SendStatusUnknown 未知
	SendStatusUnknown SendStatus = iota // 未知
	// Sending 发送中
	Sending
	// SentSuccess 发送成功
	SentSuccess
	// SendFail 发送失败
	SendFail
)
