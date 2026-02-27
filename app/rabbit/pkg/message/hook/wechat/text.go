package wechat

import "github.com/aide-family/rabbit/pkg/message"

type TextMessage struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

func NewTextMessage(content string) *TextMessage {
	return &TextMessage{
		Content: content,
	}
}

func (t *TextMessage) WithMentionedList(list []string) *TextMessage {
	t.MentionedList = list
	return t
}

func (t *TextMessage) WithMentionedMobileList(list []string) *TextMessage {
	t.MentionedMobileList = list
	return t
}

func (t *TextMessage) Message() message.Message {
	return &Message{
		MsgType: MessageTypeText,
		Text:    t,
	}
}
