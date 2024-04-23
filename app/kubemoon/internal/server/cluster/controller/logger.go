package controller

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func Logger() HandlerFunc {
	return func(c *Context) (*time.Duration, error) {
		// Start timer
		t := time.Now()
		defer log.Infow(
			"msg", "cluster handle spend",
			"key", c.Key,
			"phase", c.Phase,
			"duration", time.Since(t))
		return c.Next()
	}
}
