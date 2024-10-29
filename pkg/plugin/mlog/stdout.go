package mlog

import (
	"io"

	"github.com/go-kratos/kratos/v2/log"
)

type (
	stdoutLog struct {
		write io.Writer
		log.Logger
	}
)

// NewStdoutLogger new a stdout logger.
func NewStdoutLogger(write io.Writer) Logger {
	return &stdoutLog{
		write:  write,
		Logger: log.NewStdLogger(write),
	}
}
