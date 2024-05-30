package log

import (
	"context"
	"fmt"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/pkg/env"
	"github.com/go-kratos/kratos/v2/errors"
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

func RecoveryHandle(_ context.Context, req, err interface{}) error {
	log.Errorw("panic", err)
	myErr, ok := err.(*errors.Error)
	if ok {
		return myErr
	}
	return merr.ErrorNotification("系统错误").WithMetadata(map[string]string{
		"error":  fmt.Sprintf("%v", err),
		"params": fmt.Sprintf("%v", req),
	})
}
