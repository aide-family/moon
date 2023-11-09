package plog

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/google/wire"
)

var (
	ProviderSetPlog = wire.NewSet(
		NewLogger,
	)
)

type (
	ServerEnv interface {
		GetId() string
		GetName() string
		GetVersion() string
	}
)

// NewLogger new a logger.
func NewLogger(c ServerEnv) log.Logger {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", c.GetId(),
		"service.name", c.GetName(),
		"service.version", c.GetVersion(),
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	return logger
}
