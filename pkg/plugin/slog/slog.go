package slog

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/exp/slog"
)

var _ Logger = (*Slog)(nil)

// NewSlog returns a new slog logger.
func NewSlog() Logger {
	return &Slog{
		log:    slog.With(),
		msgKey: "msg",
	}
}

type Slog struct {
	log    *slog.Logger
	msgKey string
}

func (s *Slog) Log(level log.Level, keyvals ...interface{}) error {
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

func (s *Slog) Sync() error {
	return nil
}
