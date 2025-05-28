package gorm

import (
	"context"
	"time"

	"gorm.io/gorm/logger"

	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/go-kratos/kratos/v2/log"
)

func NewLogger(logger log.Logger) logger.Interface {
	return &gormLogger{helper: log.NewHelper(log.With(logger, "module", "gorm"))}
}

type gormLogger struct {
	helper *log.Helper
}

// Error implements logger.Interface.
func (g *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	g.helper.WithContext(ctx).Errorw("msg", msg, "data", data)
}

// Info implements logger.Interface.
func (g *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	g.helper.WithContext(ctx).Infow("msg", msg, "data", data)
}

// LogMode implements logger.Interface.
func (g *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return g
}

// Trace implements logger.Interface.
func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	duration := time.Since(begin)
	if err != nil {
		g.helper.WithContext(ctx).Errorw("begin", timex.Format(begin), "sql", sql, "rowsAffected", rowsAffected, "err", err, "duration", duration)
	} else {
		g.helper.WithContext(ctx).Debugw("begin", timex.Format(begin), "sql", sql, "rowsAffected", rowsAffected, "duration", duration)
	}
}

// Warn implements logger.Interface.
func (g *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	g.helper.WithContext(ctx).Warnw("msg", msg, "data", data)
}
