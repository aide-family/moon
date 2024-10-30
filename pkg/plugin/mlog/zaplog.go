package mlog

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = (*zapLogger)(nil)

type (
	ZapLogConfig interface {
		GetJson() bool
	}

	zapLogger struct {
		log    *zap.Logger
		msgKey string
	}
)

// NewZapLogger new a zap logger.
func NewZapLogger(c ZapLogConfig) Logger {
	zapLog := zap.NewExample()
	if c.GetJson() {
		encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
		zapLog = zap.New(zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.DebugLevel))
	}

	l := &zapLogger{
		log:    zapLog,
		msgKey: log.DefaultMessageKey,
	}

	return l
}

// Log 实现kratos log接口
func (l *zapLogger) Log(level log.Level, keyvals ...interface{}) error {
	var (
		msg    = ""
		keyLen = len(keyvals)
	)
	if keyLen == 0 {
		l.log.Warn(fmt.Sprintf("Keyvalues must appear in pairs: %v", keyvals))
		return nil
	}

	if keyLen%2 != 0 {
		msg = fmt.Sprintf("%v", keyvals[len(keyvals)-1])
		keyLen -= 1
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
