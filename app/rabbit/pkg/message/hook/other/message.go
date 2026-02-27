package other

import (
	"encoding/json"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/pkg/message"
)

var _ message.Message = (*Message)(nil)

type Message map[string]any

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Type() enum.MessageType {
	return enum.MessageType_WEBHOOK_OTHER
}
