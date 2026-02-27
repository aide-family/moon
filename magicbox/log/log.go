// Package log is a simple package that provides a logger interface.
package log

import (
	"github.com/go-kratos/kratos/v2/log"
)

type Driver interface {
	New() (Interface, error)
}

type Interface interface {
	log.Logger
}

// NewLogger create logger by  driver and config
func NewLogger(driver Driver) (logger Interface, err error) {
	return driver.New()
}
