package storage

import (
	"github.com/moon-monitor/moon/pkg/plugin"
)

func LoadPlugin(config *plugin.LoadConfig) (FileManager, error) {
	return plugin.Load[FileManager](config)
}
