package schema

import (
	"fmt"
	"regexp"

	"github.com/glebarez/sqlite"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aide-family/jade_tree/internal/data/impl/do"
)

var dsn string

var sqliteIndexAlreadyExistsRe = regexp.MustCompile(`(?i)index\s+([a-zA-Z0-9_]+)\s+already exists`)

func autoMigrateSQLiteWithIndexRetry(db *gorm.DB) error {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if err := db.AutoMigrate(do.Models()...); err == nil {
			return nil
		} else {
			lastErr = err
		}

		matches := sqliteIndexAlreadyExistsRe.FindStringSubmatch(lastErr.Error())
		if len(matches) != 2 {
			return lastErr
		}
		indexName := matches[1]
		if execErr := db.Exec("DROP INDEX IF EXISTS `" + indexName + "`").Error; execErr != nil {
			return fmt.Errorf("drop existing sqlite index %s: %w", indexName, execErr)
		}
	}
	return lastErr
}

func newMigrateCmd() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the database",
		Long:  "Migrate the database to the latest version",
	}
	migrateCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug mode")
	migrateCmd.PersistentFlags().StringVar(&dsn, "dsn", "", `Database DSN.
Example:
- sqlite: file::memory:?cache=shared
- mysql: root:123456@tcp(localhost:3306)/jade_tree?charset=utf8mb4&parseTime=True&loc=Local
- postgres: host=localhost user=root password=123456 port=5432 dbname=jade_tree sslmode=disable
	`)
	migrateCmd.AddCommand(newSQLiteMigrateCmd(), newMySQLMigrateCmd(), newPostgresMigrateCmd())
	return migrateCmd
}

func newSQLiteMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sqlite",
		Short: "Migrate the database to the latest version",
		Long:  "Migrate the database to the latest version using SQLite",
		RunE: func(c *cobra.Command, args []string) error {
			if dsn == "" {
				dsn = "file::memory:?cache=shared"
			}
			db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open sqlite: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if err := autoMigrateSQLiteWithIndexRetry(db); err != nil {
				return fmt.Errorf("migrate models: %w", err)
			}
			return nil
		},
	}
	return cmd
}

func newMySQLMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: "Migrate the database to the latest version using MySQL",
		Long:  "Migrate the database to the latest version using MySQL",
		RunE: func(c *cobra.Command, args []string) error {
			if dsn == "" {
				dsn = "root:123456@tcp(localhost:3306)/jade_tree?charset=utf8mb4&parseTime=True&loc=Local"
			}
			db, err := gorm.Open(mysql.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open mysql: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if err := db.AutoMigrate(do.Models()...); err != nil {
				return fmt.Errorf("migrate models: %w", err)
			}
			return nil
		},
	}
	return cmd
}

func newPostgresMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "postgres",
		Short: "Migrate the database to the latest version using Postgres",
		Long:  "Migrate the database to the latest version using Postgres",
		RunE: func(c *cobra.Command, args []string) error {
			if dsn == "" {
				dsn = "host=localhost user=root password=123456 port=5432 dbname=jade_tree sslmode=disable TimeZone=Asia/Shanghai"
			}
			db, err := gorm.Open(postgres.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open postgres: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if err := db.AutoMigrate(do.Models()...); err != nil {
				return fmt.Errorf("migrate models: %w", err)
			}
			return nil
		},
	}
	return cmd
}
