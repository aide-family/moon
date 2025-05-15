package hook

import "github.com/moon-monitor/moon/pkg/util/template"

type Message []byte

// FormatMessage formats the message
func FormatMessage(payload string, data any) (Message, error) {
	tpl, err := template.TextFormatter(payload, data)
	if err != nil {
		return nil, err
	}
	return []byte(tpl), nil
}
