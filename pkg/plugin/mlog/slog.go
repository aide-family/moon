package mlog

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/exp/slog"
)

var _ Logger = (*sLogger)(nil)

// NewSlog returns a new slog logger.
func NewSlog(c SLogConfig) Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}
	s := slog.New(slog.NewTextHandler(os.Stdout, opts))
	if c.GetJson() {
		s = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return &sLogger{
		log:    s,
		msgKey: "msg",
	}
}

type (
	SLogConfig interface {
		GetJson() bool
	}

	sLogger struct {
		log    *slog.Logger
		msgKey string
	}
)

func (s *sLogger) Log(level log.Level, keyvals ...interface{}) error {
	var (
		msg    = ""
		keyLen = len(keyvals)
	)
	if keyLen == 0 || keyLen%2 != 0 {
		s.log.Warn(fmt.Sprintf("Keyvalues must appear in pairs: %v", keyvals))
		return nil
	}

	data := make([]slog.Attr, 0, (keyLen/2)+1)
	for i := 0; i < keyLen; i += 2 {
		if keyvals[i].(string) == s.msgKey {
			msg, _ = keyvals[i+1].(string)
			continue
		}
		data = append(data, slog.Any(fmt.Sprint(keyvals[i]), keyvals[i+1]))
	}
	ctx := context.Background()
	switch level {
	case log.LevelDebug:
		s.log.LogAttrs(ctx, slog.LevelDebug, msg, data...)
	case log.LevelInfo:
		s.log.LogAttrs(ctx, slog.LevelInfo, msg, data...)
	case log.LevelWarn:
		s.log.LogAttrs(ctx, slog.LevelWarn, msg, data...)
	case log.LevelError:
		s.log.LogAttrs(ctx, slog.LevelError, msg, data...)
	case log.LevelFatal:
		s.log.LogAttrs(ctx, slog.LevelError, msg, data...)
	}
	return nil
}
