package conn

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/logger"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// gormContextTxKey GORM事务的上下文
type gormContextTxKey struct{}

// GormDBConfig GORM数据库配置
type GormDBConfig interface {
	GetDriver() string
	GetDsn() string
	GetDebug() bool
}

// GetDB 获取数据库连接
func GetDB(ctx context.Context) (*gorm.DB, bool) {
	if types.IsNil(ctx) {
		return nil, false
	}
	if v, ok := ctx.Value(gormContextTxKey{}).(*gorm.DB); ok {
		return v, true
	}
	return nil, false
}

// NewGormDB 获取数据库连接
func NewGormDB(c GormDBConfig, logger ...log.Logger) (*gorm.DB, error) {
	var opts []gorm.Option
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	if len(logger) > 0 {
		gormLog := NewGormLogger(logger[0])
		gormConfig.Logger = gormLog
	}
	opts = append(opts, gormConfig)

	var dialector gorm.Dialector
	dsn := c.GetDsn()
	drive := c.GetDriver()
	switch drive {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		// 判断文件是否存在，不存在则创建
		//if err := checkDBFileExists(dsn); err != nil {
		//	return nil, err
		//}
		dialector = sqlite.Open(dsn)
	default:
		return nil, merr.ErrorNotification("invalid driver: %s", drive)
	}

	conn, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, merr.ErrorNotification("connect db error: %s", err)
	}

	if drive == "sqlite" {
		// 启用 WAL 模式
		_ = conn.Exec("PRAGMA journal_mode=WAL;")
	}

	if c.GetDebug() {
		conn = conn.Debug()
	}

	return conn, nil
}

// GormLogger gorm日志实现
type GormLogger struct {
	logger log.Logger
	level  logger.LogLevel
}

// NewGormLogger Gorm日志实现
func NewGormLogger(logger log.Logger) logger.Interface {
	return &GormLogger{logger: logger}
}

// LogMode 设置日志等级
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

// Info log info
func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	_ = log.WithContext(ctx, l.logger).Log(log.LevelInfo, fmt.Sprintf(s, i...))
}

// Warn log warn
func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	_ = log.WithContext(ctx, l.logger).Log(log.LevelWarn, fmt.Sprintf(s, i...))
}

// Error log error
func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	_ = log.WithContext(ctx, l.logger).Log(log.LevelError, fmt.Sprintf(s, i...))
}

// Trace log trace
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level >= logger.Info {
		ctx, span := otel.Tracer("gorm").Start(ctx, "gorm.trace")
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
			_ = log.WithContext(ctx, l.logger).Log(log.LevelError, "sql", sql, "elapsed", elapsed, "err", err)
		} else {
			_ = log.WithContext(ctx, l.logger).Log(log.LevelInfo, "sql", sql, "elapsed", elapsed, "rows", rows)
		}
	}
}
