package conf

import (
	"github.com/aide-family/moon/pkg/config"
)

func (c *Bootstrap) IsDev() bool {
	return c.GetEnvironment() == config.Environment_DEV
}
