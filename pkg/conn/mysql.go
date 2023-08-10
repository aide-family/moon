package conn

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type DBConfig interface {
	GetDriver() string
	GetSource() string
	GetDebug() bool
}

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
			_ = log.WithContext(ctx, l.logger).Log(log.LevelError, "exec", fmt.Sprintf("[%s] %s, %v", sql, elapsed, err))
		} else {
			_ = log.WithContext(ctx, l.logger).Log(log.LevelInfo, "exec", fmt.Sprintf("[%s] %s, %d rows affected", sql, elapsed, rows))
		}
	}
}

// NewMysqlDB 获取mysql数据库连接
func NewMysqlDB(cfg DBConfig, logger ...log.Logger) (*gorm.DB, error) {
	var opts []gorm.Option
	if len(logger) > 0 {
		//gormLog := NewGormLogger(logger[0])
		//opts = append(opts, &gorm.Config{Logger: gormLog, DisableForeignKeyConstraintWhenMigrating: true})
	}

	conn, err := gorm.Open(mysql.Open(cfg.GetSource()), opts...)
	if err != nil {
		return nil, err
	}

	if cfg.GetDebug() {
		conn = conn.Debug()
	}

	return conn, nil
}
