package wechat

import "github.com/aide-family/rabbit/pkg/message"

type ImageMessage struct {
	Base64 string `json:"base64"`
	MD5    string `json:"md5"`
}

func NewImageMessage(base64 string, md5 string) *ImageMessage {
	return &ImageMessage{
		Base64: base64,
		MD5:    md5,
	}
}

func (m *ImageMessage) Message() message.Message {
	return &Message{
		MsgType: MessageTypeImage,
		Image:   m,
	}
}
