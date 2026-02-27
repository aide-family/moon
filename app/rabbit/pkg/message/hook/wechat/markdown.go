package wechat

import "github.com/aide-family/rabbit/pkg/message"

type MarkdownMessage struct {
	Content string `json:"content"`
}

func NewMarkdownMessage(content string) *MarkdownMessage {
	return &MarkdownMessage{
		Content: content,
	}
}

func (m *MarkdownMessage) Message() message.Message {
	return &Message{
		MsgType:  MessageTypeMarkdown,
		Markdown: m,
	}
}
