package schema

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func NewSQLiteSchema(db *gorm.DB) Schema {
	return &sqliteSchema{db: db}
}

type sqliteSchema struct {
	db *gorm.DB
}

func (s *sqliteSchema) CreateTableSQL(tableName string) (string, error) {
	var createSQL string
	err := s.db.Raw("SELECT sql FROM sqlite_master WHERE type = ? AND name = ?", "table", tableName).
		Scan(&createSQL).Error
	if err != nil {
		return "", fmt.Errorf("get create table sql for %s: %w", tableName, err)
	}
	if createSQL == "" {
		return "", fmt.Errorf("table not found: %s", tableName)
	}
	if !strings.HasSuffix(createSQL, ";") {
		createSQL += ";"
	}
	return createSQL + "\n", nil
}

// IndexSQL returns CREATE INDEX / CREATE UNIQUE INDEX for the table (excludes sqlite_autoindex_*).
func (s *sqliteSchema) IndexSQL(tableName string) ([]string, error) {
	var sqls []string
	tx := s.db.Raw(
		"SELECT sql FROM sqlite_master WHERE type = ? AND tbl_name = ? AND sql IS NOT NULL AND name NOT LIKE ? ORDER BY name",
		"index", tableName, "sqlite_autoindex_%",
	).Pluck("sql", &sqls)
	if tx.Error != nil {
		return nil, fmt.Errorf("list indexes for %s: %w", tableName, tx.Error)
	}
	out := make([]string, 0, len(sqls))
	for _, sql := range sqls {
		if sql == "" {
			continue
		}
		if !strings.HasSuffix(sql, ";") {
			sql += ";"
		}
		out = append(out, sql+"\n")
	}
	return out, nil
}

func (s *sqliteSchema) Tables() ([]string, error) {
	var tables []string
	tx := s.db.Raw("SELECT name FROM sqlite_master WHERE type = ? AND name NOT LIKE ? ORDER BY name", "table", "sqlite_%").
		Pluck("name", &tables)
	if tx.Error != nil {
		return nil, fmt.Errorf("list tables: %w", tx.Error)
	}
	return tables, nil
}
