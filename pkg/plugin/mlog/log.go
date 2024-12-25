package mlog

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

// New a logger.
func New(c *conf.Log) (l Logger) {
	defer func() {
		l = NewLogger(log.NewFilter(l, log.FilterLevel(log.ParseLevel(c.GetLevel()))))
		log.SetLogger(l)
	}()
	switch strings.ToLower(c.GetType()) {
	case "aliyun":
		return NewAliYunLog(c.GetAliyun())
	case "slog":
		return NewSlog(c.GetSlog())
	case "zap", "zaplog":
		return NewZapLogger(c.GetZap())
	case "loki":
		return NewLokiLogger(c.GetLoki())
	default:
		return NewLogger(NewStdoutLogger(os.Stdout))
	}
}

// NewHelper new a log helper.
func NewHelper(logger Logger, module, domain string) *log.Helper {
	return log.NewHelper(log.With(logger, "module", module, "domain", domain))
}

type (
	// Logger 日志
	Logger interface {
		log.Logger
	}

	_logger struct {
		log log.Logger
	}
)

func (l *_logger) Log(level log.Level, keyvals ...interface{}) error {
	return l.log.Log(level, keyvals...)
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
	// return log.NewStdLogger(os.Stdout)
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
	return &_logger{log: ll}
}
