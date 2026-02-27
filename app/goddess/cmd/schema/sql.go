package schema

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	schematool "github.com/aide-family/magicbox/connect/schema"
	authmodel "github.com/aide-family/magicbox/domain/auth/v1/gormimpl/model"
	membermodel "github.com/aide-family/magicbox/domain/member/v1/gormimpl/model"
	namespacemodel "github.com/aide-family/magicbox/domain/namespace/v1/gormimpl/model"
	"github.com/glebarez/sqlite"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/data/impl/do"
)

var (
	onlyDB bool
	debug  bool
	output = "deploy/sql/"

	gormConfig = &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
)

// models returns all models in migration order: namespace and auth first (referenced by do models), then do models.
func models() []any {
	models := append(namespacemodel.Models(), authmodel.Models()...)
	models = append(models, membermodel.Models()...)
	models = append(models, do.Models()...)
	return models
}

func newSQLCmd() *cobra.Command {
	sqlCmd := &cobra.Command{
		Use:   "sql",
		Short: "Dump table schema to a SQL file",
		Long:  "Dump DDL from GORM models. Use subcommand: sqlite (in-memory) or mysql (with --host/--user/--pass).",
	}
	sqlCmd.PersistentFlags().BoolVar(&onlyDB, "only-db", false, "Only generate schema for database")
	sqlCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug mode")
	sqlCmd.PersistentFlags().StringVarP(&output, "output", "o", output, "Output SQL file path")
	sqlCmd.AddCommand(newSQLiteCmd(), newMySQLCmd(), newPostgresCmd())
	return sqlCmd
}

func newSQLiteCmd() *cobra.Command {
	dsn := "file::memory:?cache=shared"
	cmd := &cobra.Command{
		Use:   "sqlite",
		Short: "Dump schema using in-memory SQLite",
		RunE: func(c *cobra.Command, args []string) error {
			db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open sqlite: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if !onlyDB {
				if err := db.AutoMigrate(models()...); err != nil {
					return fmt.Errorf("migrate models: %w", err)
				}
			}
			filename := "schema-sqlite.sql"
			filePath := filepath.Join(output, filename)
			if err := schematool.DumpSchema(schematool.NewSQLiteSchema(db), filePath); err != nil {
				return err
			}
			klog.Infof("Schema written to %s", filePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&dsn, "dsn", dsn, "SQLite DSN")
	return cmd
}

func newMySQLCmd() *cobra.Command {
	var host, user, pass, port, database string
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: "Dump schema from MySQL (create temp DB and migrate from models; for existing DB use: schema from-db mysql)",
		RunE: func(c *cobra.Command, args []string) error {
			if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
			}
			var dsn string
			if database == "" {
				dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
					user, url.QueryEscape(pass), host, port)
				conn, err := gorm.Open(mysql.Open(dsn), gormConfig)
				if err != nil {
					return fmt.Errorf("open mysql: %w", err)
				}
				database = "rabbit_schema_tmp"
				if err := conn.Exec("CREATE DATABASE IF NOT EXISTS `" + database + "`").Error; err != nil {
					return fmt.Errorf("create database: %w", err)
				}
				rawDB, err := conn.DB()
				if err != nil {
					return fmt.Errorf("get db: %w", err)
				}
				if err := rawDB.Close(); err != nil {
					return fmt.Errorf("close db: %w", err)
				}
				dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
					user, url.QueryEscape(pass), host, port, database)
			}

			db, err := gorm.Open(mysql.Open(dsn), gormConfig)
			if err != nil {
				return fmt.Errorf("open mysql: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if !onlyDB {
				if err := db.AutoMigrate(models()...); err != nil {
					return fmt.Errorf("migrate models: %w", err)
				}
			}
			filename := "schema-mysql.sql"
			filePath := filepath.Join(output, filename)
			if err := schematool.DumpSchema(schematool.NewMySQLSchema(db), filePath); err != nil {
				return err
			}
			klog.Infof("Schema written to %s", filePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&host, "host", "localhost", "MySQL host")
	cmd.Flags().StringVar(&user, "user", "root", "MySQL user")
	cmd.Flags().StringVar(&pass, "pass", "123456", "MySQL password")
	cmd.Flags().StringVar(&port, "port", "3306", "MySQL port")
	cmd.Flags().StringVar(&database, "database", "", "MySQL database name")
	return cmd
}

func newPostgresCmd() *cobra.Command {
	var host, user, pass, port, database string
	cmd := &cobra.Command{
		Use:   "postgres",
		Short: "Dump schema from Postgres (create temp DB and migrate from models; for existing DB use: schema from-db postgres)",
		RunE: func(c *cobra.Command, args []string) error {
			if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
			}
			var dsn string
			if database == "" {
				// Postgres requires connecting to an existing DB; use "postgres" to create temp DB.
				dsn = fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=postgres sslmode=disable", host, user, pass, port)
				conn, err := gorm.Open(postgres.Open(dsn), gormConfig)
				if err != nil {
					return fmt.Errorf("open postgres: %w", err)
				}
				database = "rabbit_schema_tmp"
				err = conn.Exec(`CREATE DATABASE "` + database + `"`).Error
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("create database: %w", err)
				}
				rawDB, err := conn.DB()
				if err != nil {
					return fmt.Errorf("get db: %w", err)
				}
				if err := rawDB.Close(); err != nil {
					return fmt.Errorf("close db: %w", err)
				}
				dsn = fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=disable", host, user, pass, port, database)
			}
			// PrepareStmt: true avoids pgx "insufficient arguments" in Migrator.GetRows (SELECT * FROM table LIMIT 1).
			postgresConfig := *gormConfig
			postgresConfig.PrepareStmt = true
			db, err := gorm.Open(postgres.Open(dsn), &postgresConfig)
			if err != nil {
				return fmt.Errorf("open postgres: %w", err)
			}
			if debug {
				db = db.Debug()
			}
			if !onlyDB {
				// Drop tables first so we always create from scratch.
				_ = db.Migrator().DropTable(models()...)
				if err := db.AutoMigrate(models()...); err != nil {
					return fmt.Errorf("migrate models: %w", err)
				}
			}
			filename := "schema-postgres.sql"
			filePath := filepath.Join(output, filename)
			if err := schematool.DumpSchema(schematool.NewPostgresSchema(db), filePath); err != nil {
				return err
			}
			klog.Infof("Schema written to %s", filePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&host, "host", "localhost", "Postgres host")
	cmd.Flags().StringVar(&user, "user", "root", "Postgres user")
	cmd.Flags().StringVar(&pass, "pass", "123456", "Postgres password")
	cmd.Flags().StringVar(&port, "port", "5432", "Postgres port")
	cmd.Flags().StringVar(&database, "database", "", "Postgres database name")
	return cmd
}
