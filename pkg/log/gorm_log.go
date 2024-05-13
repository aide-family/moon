package log

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	logger log.Logger
	level  logger.LogLevel
}

func NewGormLogger(logger log.Logger) logger.Interface {
	return &GormLogger{logger: logger}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	_ = log.WithContext(ctx, l.logger).Log(log.LevelInfo, fmt.Sprintf(s, i...))
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	_ = log.WithContext(ctx, l.logger).Log(log.LevelWarn, fmt.Sprintf(s, i...))
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	_ = log.WithContext(ctx, l.logger).Log(log.LevelError, fmt.Sprintf(s, i...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level >= logger.Info {
		ctx, span := otel.Tracer("gorm").Start(ctx, "mysql.gorm.trace")
		defer span.End()

		elapsed := time.Since(begin)
		sql, rows := fc()

		span.SetAttributes(
			attribute.String("sql", sql),
			attribute.Int64("rows", rows),
			attribute.String("elapsed", elapsed.String()),
			attribute.String("error", fmt.Sprintf("%v", err)),
		)

		if err != nil {
			_ = log.WithContext(ctx, l.logger).Log(log.LevelError, "sql", sql, "elapsed", elapsed, "err", elapsed)
		} else {
			_ = log.WithContext(ctx, l.logger).Log(log.LevelInfo, "sql", sql, "elapsed", elapsed, "rows", rows)
		}
	}
}
