package conf

import (
	"github.com/moon-monitor/moon/pkg/config"
)

func (c *Bootstrap) IsDev() bool {
	return c.GetEnvironment() == config.Environment_DEV
}
