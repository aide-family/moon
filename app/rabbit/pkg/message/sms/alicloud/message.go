package alicloud

import (
	"encoding/json"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/pkg/message"
)

var _ message.Message = (*Message)(nil)

type Message struct {
	TemplateParam string   `json:"templateParam"`
	TemplateCode  string   `json:"templateCode"`
	PhoneNumbers  []string `json:"phoneNumbers"`
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Type() enum.MessageType {
	return enum.MessageType_SMS_ALICLOUD
}
