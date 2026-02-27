package schema

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func NewMySQLSchema(db *gorm.DB) Schema {
	return &mysqlSchema{db: db}
}

type mysqlSchema struct {
	db *gorm.DB
}

// escapeMySQLIdentifier escapes backticks in table name for use in SHOW CREATE TABLE.
func escapeMySQLIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// IndexSQL returns nil: MySQL includes all indexes in SHOW CREATE TABLE.
func (s *mysqlSchema) IndexSQL(tableName string) ([]string, error) {
	return nil, nil
}

func (s *mysqlSchema) CreateTableSQL(tableName string) (string, error) {
	var name, createSQL string
	query := fmt.Sprintf("SHOW CREATE TABLE %s", escapeMySQLIdentifier(tableName))
	row := s.db.Raw(query).Row()
	if err := row.Scan(&name, &createSQL); err != nil {
		return "", fmt.Errorf("show create table %s: %w", tableName, err)
	}
	if createSQL != "" && !strings.HasSuffix(createSQL, ";") {
		createSQL += ";"
	}
	return createSQL + "\n", nil
}

func (s *mysqlSchema) Tables() ([]string, error) {
	var tables []string
	tx := s.db.Raw("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() ORDER BY TABLE_NAME").
		Pluck("TABLE_NAME", &tables)
	if tx.Error != nil {
		return nil, fmt.Errorf("list tables: %w", tx.Error)
	}
	return tables, nil
}
