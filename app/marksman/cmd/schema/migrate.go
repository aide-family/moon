package schema

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/cmd"
)

var dsn string

func newMigrateCmd() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the database",
		Long:  "Migrate the database to the latest version",
		Annotations: map[string]string{
			"group": cmd.DatabaseCommands,
		},
	}
	migrateCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug mode")
	migrateCmd.PersistentFlags().StringVar(&dsn, "dsn", "", `Database DSN.
Example:
- sqlite: file::memory:?cache=shared
- mysql: root:123456@tcp(localhost:3306)/marksman?charset=utf8mb4&parseTime=True&loc=Local
- postgres: host=localhost user=root password=123456 port=5432 dbname=marksman sslmode=disable
	`)
	migrateCmd.AddCommand(newSQLiteMigrateCmd(), newMySQLMigrateCmd(), newPostgresMigrateCmd())
	return migrateCmd
}

func newSQLiteMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sqlite",
		Short: "Migrate the database to the latest version",
		Long:  "Migrate the database to the latest version using SQLite",
		Annotations: map[string]string{
			"group": cmd.DatabaseCommands,
		},
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
			if err := db.AutoMigrate(models()...); err != nil {
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
		Annotations: map[string]string{
			"group": cmd.DatabaseCommands,
		},
		RunE: func(c *cobra.Command, args []string) error {
			if dsn == "" {
				dsn = "root:123456@tcp(localhost:3306)/marksman?charset=utf8mb4&parseTime=True&loc=Local"
			}
			db, err := gorm.Open(mysql.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open mysql: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if err := db.AutoMigrate(models()...); err != nil {
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
		Annotations: map[string]string{
			"group": cmd.DatabaseCommands,
		},
		RunE: func(c *cobra.Command, args []string) error {
			if dsn == "" {
				dsn = "host=localhost user=root password=123456 port=5432 dbname=marksman sslmode=disable TimeZone=Asia/Shanghai"
			}
			db, err := gorm.Open(postgres.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open postgres: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if err := db.AutoMigrate(models()...); err != nil {
				return fmt.Errorf("migrate models: %w", err)
			}
			return nil
		},
	}
	return cmd
}
