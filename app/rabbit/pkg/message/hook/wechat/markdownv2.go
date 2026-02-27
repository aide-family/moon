package wechat

import "github.com/aide-family/rabbit/pkg/message"

type MarkdownV2Message struct {
	Content string `json:"content"`
}

func NewMarkdownV2Message(content string) *MarkdownV2Message {
	return &MarkdownV2Message{
		Content: content,
	}
}

func (m *MarkdownV2Message) Message() message.Message {
	return &Message{
		MsgType:    MessageTypeMarkdownV2,
		MarkdownV2: m,
	}
}
