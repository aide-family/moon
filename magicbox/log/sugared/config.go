package sugared

import (
	"github.com/go-kratos/kratos/v2/log"
)

// Config is a config for the logger.
type Config interface {
	GetLevel() log.Level
	GetFormat() Formatter
	GetOutput() string
	GetEnableCaller() bool
	GetEnableColor() bool
	GetEnableStack() bool
	IsDev() bool
}

// Formatter is a formatter for the logger.
type Formatter string

const (
	FormatterConsole Formatter = "console"
	FormatterJSON    Formatter = "json"
)
