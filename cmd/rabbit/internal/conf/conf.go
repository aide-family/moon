// Package conf is a configuration package for kratos.
package conf

import (
	"github.com/aide-family/moon/pkg/config"
)

func (c *Bootstrap) IsDev() bool {
	return c.GetEnvironment() == config.Environment_DEV
}
