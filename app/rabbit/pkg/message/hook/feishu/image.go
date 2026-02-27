package feishu

import "github.com/aide-family/rabbit/pkg/message"

func NewImageMessage(imageKey string) message.Message {
	return &Message{
		MsgType: MessageTypeImage,
		Content: &Content{
			Image: imageKey,
		},
	}
}
