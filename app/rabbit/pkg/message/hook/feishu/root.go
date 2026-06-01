package feishu

import (
	"encoding/json"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/pkg/message"
)

var _ message.Message = (*Message)(nil)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypePost  MessageType = "post"
	MessageTypeImage MessageType = "image"
	MessageTypeCard  MessageType = "interactive"
)

type Content struct {
	Text  *Text  `json:"text,omitempty"`
	Post  *Post  `json:"post,omitempty"`
	Card  *Card  `json:"card,omitempty"`
	Image string `json:"image_key,omitempty"`
}

type Message struct {
	MsgType   MessageType `json:"msg_type"`
	Content   *Content    `json:"content"`
	Timestamp string      `json:"timestamp,omitempty"`
	Sign      string      `json:"sign,omitempty"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Type() enum.MessageType {
	return enum.MessageType_WEBHOOK_FEISHU
}

func (m *Message) Signature(secret string) error {
	sign, err := feishuWebhookSign(m.Timestamp, secret)
	if err != nil {
		return err
	}
	m.Sign = sign
	return nil
}
