package conn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
			_ = log.WithContext(ctx, l.logger).Log(log.LevelError, "sql", sql, "elapsed", elapsed, "err", elapsed)
		} else {
			_ = log.WithContext(ctx, l.logger).Log(log.LevelInfo, "sql", sql, "elapsed", elapsed, "rows", rows)
		}
	}
}

// NewDB 获取数据库连接
func NewDB(cfg DBConfig, logger ...log.Logger) (*gorm.DB, error) {
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
	switch cfg.GetDriver() {
	case "mysql":
		dialector = mysql.Open(cfg.GetSource())
	case "sqlite":
		// 判断文件是否存在，不存在则创建
		if err := checkDBFileExists(cfg.GetSource()); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.GetSource())

	default:
		return nil, fmt.Errorf("invalid driver: %s", cfg.GetDriver())
	}

	conn, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, err
	}

	if cfg.GetDriver() == "sqlite" {
		// 启用 WAL 模式
		_ = conn.Exec("PRAGMA journal_mode=WAL;")
	}

	if cfg.GetDebug() {
		conn = conn.Debug()
	}

	return conn, nil
}

// checkDBFileExists .
func checkDBFileExists(filename string) error {
	log.Debugw("-------------------------", filename)
	if filename == "" {
		return fmt.Errorf("db file is empty")
	}
	file, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建文件夹
			dir := filepath.Dir(filename)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
			// 创建文件
			f, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer f.Close()
			return nil
		}
	}
	if file.IsDir() {
		return fmt.Errorf("db file is dir")
	}
	return err
}
