package log

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
)

var _ log.Logger = (*zapLogger)(nil)

type zapLogger struct {
	log    *zap.Logger
	msgKey string
}

type Option func(*zapLogger)

// WithMessageKey with message key.
func WithMessageKey(key string) Option {
	return func(l *zapLogger) {
		l.msgKey = key
	}
}

// WithZapLogger with zap logger.
func WithZapLogger(zl *zap.Logger) Option {
	return func(l *zapLogger) {
		l.log = zl
	}
}

func NewLogger(opts ...Option) *zapLogger {
	l := &zapLogger{
		log:    zap.NewExample(),
		msgKey: log.DefaultMessageKey,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *zapLogger) Log(level log.Level, keyvals ...interface{}) error {
	var (
		msg    = ""
		keyLen = len(keyvals)
	)
	if keyLen == 0 || keyLen%2 != 0 {
		l.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}

	data := make([]zap.Field, 0, (keyLen/2)+1)
	for i := 0; i < keyLen; i += 2 {
		if keyvals[i].(string) == l.msgKey {
			msg, _ = keyvals[i+1].(string)
			continue
		}
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), keyvals[i+1]))
	}

	switch level {
	case log.LevelDebug:
		l.log.Debug(msg, data...)
	case log.LevelInfo:
		l.log.Info(msg, data...)
	case log.LevelWarn:
		l.log.Warn(msg, data...)
	case log.LevelError:
		l.log.Error(msg, data...)
	case log.LevelFatal:
		l.log.Fatal(msg, data...)
	}
	return nil
}

func (l *zapLogger) Sync() error {
	return l.log.Sync()
}

func (l *zapLogger) Close() error {
	return l.Sync()
}
