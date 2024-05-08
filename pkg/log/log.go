package log

import (
	"github.com/aide-cloud/moon/pkg/env"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

var logger = log.With(NewLogger(),
	"ts", log.DefaultTimestamp,
	"caller", log.DefaultCaller,
	"service.id", env.ID(),
	"service.name", env.Name(),
	"service.version", env.Version(),
	"trace.id", tracing.TraceID(),
	"span.id", tracing.SpanID(),
)

// GetLogger 获取日志实例
func GetLogger() log.Logger {
	return logger
}
