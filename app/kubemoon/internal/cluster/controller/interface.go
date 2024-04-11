package controller

import "time"

type Interface interface {
	Initial(c *Context) (*time.Duration, error)

	Running(c *Context) (*time.Duration, error)

	// Terminating will release this cluster.
	Terminating(c *Context) (*time.Duration, error)
}
