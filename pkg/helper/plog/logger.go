package plog

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/google/wire"
	"prometheus-manager/pkg/util/hello"
)

var (
	ProviderSetPLog = wire.NewSet(
		NewLogger,
	)
)

// NewLogger new a logger.
func NewLogger() log.Logger {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", hello.ID(),
		"service.name", hello.Name(),
		"service.version", hello.Version(),
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	return logger
}
