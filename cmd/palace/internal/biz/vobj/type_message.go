package vobj

// MessageType message type
//
//go:generate stringer -type=MessageType -linecomment -output=type_message.string.go
type MessageType int8

const (
	MessageTypeUnknown      MessageType = iota // unknown
	MessageTypeEmail                           // email
	MessageTypeSMS                             // sms
	MessageTypeVoice                           // voice
	MessageTypeHookDingTalk                    // dingtalk
	MessageTypeHookWechat                      // wechat
	MessageTypeHookFeishu                      // feishu
	MessageTypeHookWebhook                     // webhook
)
