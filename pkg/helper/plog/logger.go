package plog

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"prometheus-manager/pkg/util/hello"
)

var (
	ProviderSetPLog            = wire.NewSet(NewLogger)
	_               log.Logger = (*pLogger)(nil)
)

type pLogger struct {
	*zap.Logger
}

func isNil(v any) bool {
	return v == nil || fmt.Sprintf("%v", v) == "<nil>"
}

// NewLogger new a logger.
func NewLogger(c Config) log.Logger {
	l := log.NewStdLogger(os.Stdout)
	if !isNil(c) {
		l = &pLogger{
			Logger: NewZapLog(c),
		}
	}

	return log.With(l,
		"caller", log.DefaultCaller,
		"service.id", hello.ID(),
		"service.name", hello.Name(),
		"service.version", hello.Version(),
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
}

func (l *pLogger) Log(level log.Level, keyvals ...any) error {
	fields := make([]zapcore.Field, 0, len(keyvals)/2)
	var msg string
	if len(keyvals) > 0 {
		kvs := keyvals
		if len(kvs)%2 != 0 {
			kvs = append(kvs, "-")
		}
		for i := 0; i < len(kvs); i += 2 {
			fields = append(fields, zap.Any(kvs[i].(string), kvs[i+1]))
		}
	}

	switch level {
	case log.LevelDebug:
		l.Debug(msg, fields...)
	case log.LevelInfo:
		l.Info(msg, fields...)
	case log.LevelWarn:
		l.Warn(msg, fields...)
	case log.LevelError:
		l.Error(msg, fields...)
	case log.LevelFatal:
		l.Fatal(msg, fields...)
	default:
		l.Info(msg, fields...)
	}

	return nil
}
