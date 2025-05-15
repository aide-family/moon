package gorm

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/merr"
)

type DB interface {
	GetDB() *gorm.DB
	Close() error
}

// NewDB creates a new DB instance
func NewDB(c *config.Database) (DB, error) {
	// check db name exist, if not, create it
	if c.GetDbName() == "" {
		return nil, merr.ErrorBadRequest("db name is empty")
	}
	if c.GetDbName() != "" {
		sqlDB, err := newSqlDB(c)
		if err != nil {
			panic(err)
		}
		defer sqlDB.Close()
		if _, err := sqlDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;", c.GetDbName())); err != nil {
			panic(err)
		}
	}

	var opts []gorm.Option
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	opts = append(opts, gormConfig)
	var dialector gorm.Dialector
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.GetUser(), c.GetPassword(), c.GetHost(), c.GetPort(), c.GetDbName(), c.GetParams())
	drive := c.GetDriver()
	switch drive {
	case config.Database_MYSQL:
		dialector = mysql.Open(dsn)
	default:
		return nil, merr.ErrorInternalServerError("invalid driver: %s", drive)
	}
	conn, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, merr.ErrorInternalServerError("connect db error: %s", err)
	}
	if c.GetDebug() {
		conn = conn.Debug()
	}
	return &db{DB: conn}, nil
}

type db struct {
	*gorm.DB
}

// GetDB This method checks if there is a transaction in the context,
// and if so returns the client with the transaction
func (t *db) GetDB() *gorm.DB {
	return t.DB
}

// Close This method closes the DB instance
func (t *db) Close() error {
	s, err := t.DB.DB()
	if err != nil {
		return err
	}
	return s.Close()
}

func newSqlDB(c *config.Database) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.GetUser(), c.GetPassword(), c.GetHost(), c.GetPort())
	switch c.GetDriver() {
	case config.Database_MYSQL:
		return sql.Open("mysql", dsn)
	default:
		return nil, merr.ErrorInternalServerError("invalid driver: %s", c.GetDriver())
	}
}
