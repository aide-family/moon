package feishu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
	Timestamp string      `json:"timestamp"`
	Sign      string      `json:"sign"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Type() enum.MessageType {
	return enum.MessageType_WEBHOOK_FEISHU
}

func (m *Message) Signature(secret string) error {
	// 正确拼接 timestamp 和 secret，使用换行符分隔
	signString := m.Timestamp + "\n" + secret

	h := hmac.New(sha256.New, []byte(signString))
	var data []byte
	_, err := h.Write(data)
	if err != nil {
		return err
	}

	m.Sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return nil
}
