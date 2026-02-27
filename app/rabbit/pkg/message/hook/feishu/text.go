package feishu

import (
	"strings"

	"github.com/aide-family/rabbit/pkg/message"
)

type Text string

func NewTextMessage(text string) *Text {
	t := Text(text)
	return &t
}

func (t *Text) At(userID, userName string) *Text {
	builder := strings.Builder{}
	builder.WriteString(string(*t))
	builder.WriteString("<at user_id=\"")
	builder.WriteString(userID)
	builder.WriteString("\">")
	builder.WriteString(userName)
	builder.WriteString("</at>")
	*t = Text(builder.String())
	return t
}

func (t *Text) AtAll() *Text {
	builder := strings.Builder{}
	builder.WriteString(string(*t))
	builder.WriteString("<at user_id=\"all\">所有人</at>")
	*t = Text(builder.String())
	return t
}

func (t *Text) Message() message.Message {
	return &Message{
		MsgType: MessageTypeText,
		Content: &Content{
			Text: t,
		},
	}
}
