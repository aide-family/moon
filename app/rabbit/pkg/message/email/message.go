package email

import (
	"encoding/json"
	"net/http"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/pkg/message"
)

var _ message.Message = (*Message)(nil)

type Attachment struct {
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

type Message struct {
	To          []string      `json:"to"`
	Cc          []string      `json:"cc"`
	Subject     string        `json:"subject"`
	Body        string        `json:"body"`
	ContentType string        `json:"contentType"`
	Attachments []*Attachment `json:"attachments"`
	Headers     http.Header   `json:"headers"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Type() enum.MessageType {
	return enum.MessageType_EMAIL
}

func NewMessage() *Message {
	return &Message{
		Headers: map[string][]string{},
	}
}

func (m *Message) AppendTo(to ...string) *Message {
	m.To = append(m.To, to...)
	return m
}

func (m *Message) AppendCc(cc ...string) *Message {
	m.Cc = append(m.Cc, cc...)
	return m
}

func (m *Message) SetSubject(subject string) *Message {
	m.Subject = subject
	return m
}

func (m *Message) SetBody(body string) *Message {
	m.Body = body
	return m
}

func (m *Message) SetAttachments(attachments ...*Attachment) *Message {
	m.Attachments = append(m.Attachments, attachments...)
	return m
}

func (m *Message) SetHeader(key string, values ...string) *Message {
	for _, value := range values {
		m.Headers.Add(key, value)
	}
	return m
}

func (m *Message) SetContentType(contentType string) *Message {
	m.ContentType = contentType
	return m
}
