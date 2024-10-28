package slog

import (
	"context"
	"fmt"

	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

type (
	Logger interface {
		log.Logger
		Sync() error
	}

	_logger struct {
		log  log.Logger
		sync func() error
	}
)

func (l *_logger) Log(level log.Level, keyvals ...interface{}) error {
	return l.log.Log(level, keyvals...)
}

func (l *_logger) Sync() error {
	return l.sync()
}

// ID 获取服务ID
func ID() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Env()
	}
}

// Name 获取服务名称
func Name() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Name()
	}
}

// Version 获取服务版本
func Version() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Version()
	}
}

// Env 获取服务环境
func Env() log.Valuer {
	return func(ctx context.Context) interface{} {
		return env.Env()
	}
}

// RecoveryHandle 错误处理
func RecoveryHandle(_ context.Context, req, err interface{}) error {
	log.Errorw("panic", err)
	myErr, ok := err.(*errors.Error)
	if ok {
		return myErr
	}
	paramsBs, _ := types.Marshal(req)
	return errors.New(500, "SYSTEM_ERR", "system panic").WithMetadata(map[string]string{
		"error":  fmt.Sprintf("%v", err),
		"params": string(paramsBs),
	})
}

// NewLogger new a logger.
func NewLogger(l Logger) Logger {
	//return log.NewStdLogger(os.Stdout)
	ll := log.With(l,
		"ts", log.DefaultTimestamp,
		"caller", log.Caller(5),
		"service.id", ID(),
		"service.name", Name(),
		"service.version", Version(),
		"service.env", Env(),
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	return &_logger{log: ll, sync: l.Sync}
}
