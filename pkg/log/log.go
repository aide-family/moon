package log

import (
	"context"

	"github.com/aide-cloud/moon/pkg/env"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

func ID() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Env()
	}
}

func Name() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Name()
	}
}

func Version() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Version()
	}
}

func Env() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Env()
	}
}

var defaultLogger = log.With(NewLogger(),
	"ts", log.DefaultTimestamp,
	"caller", log.DefaultCaller,
	"service.id", ID(),
	"service.name", Name(),
	"service.version", Version(),
	"service.env", Env(),
	"trace.id", tracing.TraceID(),
	"span.id", tracing.SpanID(),
)

// GetLogger 获取日志实例
func GetLogger() log.Logger {
	return defaultLogger
}
