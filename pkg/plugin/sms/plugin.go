package sms

import (
	"github.com/aide-family/moon/pkg/plugin"
)

func LoadPlugin(config *plugin.LoadConfig) (Sender, error) {
	return plugin.Load[Sender](config)
}
