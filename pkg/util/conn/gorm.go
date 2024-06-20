package conn

import (
	"os"
	"path/filepath"

	"github.com/aide-family/moon/api/merr"
	slog "github.com/aide-family/moon/pkg/util/log"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var debug = true

// SetDebug 设置debug
func SetDebug(b bool) {
	debug = b
}

// GormContextTxKey GORM事务的上下文
type GormContextTxKey struct{}

// NewGormDB 获取数据库连接
func NewGormDB(dsn, drive string, logger ...log.Logger) (*gorm.DB, error) {
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
	switch drive {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		// 判断文件是否存在，不存在则创建
		if err := checkDBFileExists(dsn); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(dsn)
	default:
		return nil, merr.ErrorDependencyErr("invalid driver: %s", drive)
	}

	conn, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, merr.ErrorDbConnectErr("connect db error: %s", err)
	}

	if drive == "sqlite" {
		// 启用 WAL 模式
		_ = conn.Exec("PRAGMA journal_mode=WAL;")
	}

	if debug {
		conn = conn.Debug()
	}

	return conn, nil
}

// checkDBFileExists .
func checkDBFileExists(filename string) error {
	if filename == "" {
		return merr.ErrorDependencyErr("db file is empty")
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
		return merr.ErrorDependencyErr("db file is dir")
	}
	return err
}
