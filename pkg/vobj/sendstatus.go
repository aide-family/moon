package vobj

// SendStatus 发送状态
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=SendStatus -linecomment
type SendStatus int

const (
	// SendStatusUnknown 未知
	SendStatusUnknown SendStatus = iota // 未知
	// Sending 发送中
	Sending // 发送中
	// SentSuccess 发送成功
	SentSuccess // 发送成功
	// SendFail 发送失败
	SendFail // 发送失败
)

// EnString 转换为字符串
func (s SendStatus) EnString() string {
	switch s {
	case Sending:
		return "sending"
	case SentSuccess:
		return "sent_success"
	case SendFail:
		return "send_fail"
	default:
		return "unknown"
	}
}
