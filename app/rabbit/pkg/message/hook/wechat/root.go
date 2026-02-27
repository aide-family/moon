// Package wechat provides a set of message types for WeChat.
package wechat

import (
	"encoding/json"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/pkg/message"
)

type MessageType string

const (
	MessageTypeText       MessageType = "text"
	MessageTypeMarkdown   MessageType = "markdown"
	MessageTypeMarkdownV2 MessageType = "markdown_v2"
	MessageTypeImage      MessageType = "image"
)

var _ message.Message = (*Message)(nil)

type Message struct {
	MsgType    MessageType        `json:"msgtype"`
	Text       *TextMessage       `json:"text,omitempty"`
	Markdown   *MarkdownMessage   `json:"markdown,omitempty"`
	MarkdownV2 *MarkdownV2Message `json:"markdown_v2,omitempty"`
	Image      *ImageMessage      `json:"image,omitempty"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Type() enum.MessageType {
	return enum.MessageType_WEBHOOK_WECHAT
}
