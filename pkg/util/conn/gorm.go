package conn

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/slog"
	"github.com/aide-family/moon/pkg/util/types"

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
		gormLog := slog.NewGormLogger(logger[0])
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

// checkDBFileExists .
func checkDBFileExists(filename string) error {
	if filename == "" {
		return merr.ErrorNotification("db file is empty")
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
		return merr.ErrorNotification("db file is dir")
	}
	return err
}
