package controller

import (
	"k8s.io/klog/v2"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) (*time.Duration, error) {
		// Start timer
		t := time.Now()
		defer klog.InfoS("cluster handle spend",
			"key", c.Key,
			"phase", c.Phase,
			"duration", time.Since(t))
		// Process request
		return c.Next()
	}
}
