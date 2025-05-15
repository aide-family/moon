package sms

import (
	"github.com/moon-monitor/moon/pkg/plugin"
)

func LoadPlugin(config *plugin.LoadConfig) (Sender, error) {
	return plugin.Load[Sender](config)
}
