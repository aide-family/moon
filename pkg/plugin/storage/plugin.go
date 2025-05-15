package storage

import (
	"github.com/aide-family/moon/pkg/plugin"
)

func LoadPlugin(config *plugin.LoadConfig) (FileManager, error) {
	return plugin.Load[FileManager](config)
}
