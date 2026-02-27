package schema

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func NewPostgresSchema(db *gorm.DB) Schema {
	return &postgresSchema{db: db}
}

type postgresSchema struct {
	db *gorm.DB
}

type pgColumnInfo struct {
	OrdinalPosition      int
	ColumnName           string
	DataType             string
	ColumnDefault        sql.NullString
	IsNullable           string
	CharMaxLength        sql.NullInt64
	NumericPrecision     sql.NullInt64
	NumericScale         sql.NullInt64
	DatetimePrecision    sql.NullInt64
	CharacterOctetLength sql.NullInt64
}

func (s *postgresSchema) Tables() ([]string, error) {
	var tables []string
	// current_schema() for the connected user (typically "public")
	tx := s.db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema()
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`).Pluck("table_name", &tables)
	if tx.Error != nil {
		return nil, fmt.Errorf("list tables: %w", tx.Error)
	}
	return tables, nil
}

func (s *postgresSchema) CreateTableSQL(tableName string) (string, error) {
	var cols []pgColumnInfo
	tx := s.db.Raw(`
		SELECT ordinal_position, column_name, data_type, column_default, is_nullable,
		       character_maximum_length, numeric_precision, numeric_scale, datetime_precision,
		       character_octet_length
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ?
		ORDER BY ordinal_position
	`, tableName).Scan(&cols)
	if tx.Error != nil {
		return "", fmt.Errorf("get columns for %s: %w", tableName, tx.Error)
	}
	if len(cols) == 0 {
		return "", fmt.Errorf("table not found: %s", tableName)
	}

	quotedTable := quoteIdent(tableName)
	parts := make([]string, 0, len(cols)+4)

	for _, c := range cols {
		typ := s.fullDataType(c)
		def := ""
		if c.ColumnDefault.Valid && c.ColumnDefault.String != "" {
			def = " DEFAULT " + c.ColumnDefault.String
		}
		null := ""
		if c.IsNullable == "NO" {
			null = " NOT NULL"
		}
		parts = append(parts, "  "+quoteIdent(c.ColumnName)+" "+typ+def+null)
	}

	// Primary key
	var pkCols []string
	s.db.Raw(`
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = current_schema() AND tc.table_name = ? AND tc.constraint_type = 'PRIMARY KEY'
		ORDER BY kcu.ordinal_position
	`, tableName).Pluck("column_name", &pkCols)
	if len(pkCols) > 0 {
		quoted := make([]string, len(pkCols))
		for i, c := range pkCols {
			quoted[i] = quoteIdent(c)
		}
		parts = append(parts, "  PRIMARY KEY ("+strings.Join(quoted, ", ")+")")
	}

	// Unique constraints (excluding PK)
	var uniques []struct {
		ConstraintName string
	}
	s.db.Raw(`
		SELECT constraint_name
		FROM information_schema.table_constraints
		WHERE table_schema = current_schema() AND table_name = ? AND constraint_type = 'UNIQUE'
	`, tableName).Scan(&uniques)
	for _, u := range uniques {
		var ucols []string
		s.db.Raw(`
			SELECT column_name
			FROM information_schema.key_column_usage
			WHERE table_schema = current_schema() AND table_name = ? AND constraint_name = ?
			ORDER BY ordinal_position
		`, tableName, u.ConstraintName).Pluck("column_name", &ucols)
		if len(ucols) > 0 {
			quoted := make([]string, len(ucols))
			for i, c := range ucols {
				quoted[i] = quoteIdent(c)
			}
			parts = append(parts, "  CONSTRAINT "+quoteIdent(u.ConstraintName)+" UNIQUE ("+strings.Join(quoted, ", ")+")")
		}
	}

	createSQL := "CREATE TABLE " + quotedTable + " (\n" + strings.Join(parts, ",\n") + "\n);"
	return createSQL + "\n", nil
}

func (s *postgresSchema) fullDataType(c pgColumnInfo) string {
	switch c.DataType {
	case "character varying", "varchar":
		if c.CharMaxLength.Valid && c.CharMaxLength.Int64 > 0 {
			return fmt.Sprintf("character varying(%d)", c.CharMaxLength.Int64)
		}
		return "character varying"
	case "character", "char":
		if c.CharMaxLength.Valid && c.CharMaxLength.Int64 > 0 {
			return fmt.Sprintf("character(%d)", c.CharMaxLength.Int64)
		}
		return "character(1)"
	case "numeric", "decimal":
		if c.NumericPrecision.Valid {
			if c.NumericScale.Valid && c.NumericScale.Int64 > 0 {
				return fmt.Sprintf("numeric(%d,%d)", c.NumericPrecision.Int64, c.NumericScale.Int64)
			}
			return fmt.Sprintf("numeric(%d)", c.NumericPrecision.Int64)
		}
		return "numeric"
	case "timestamp with time zone", "timestamptz":
		if c.DatetimePrecision.Valid && c.DatetimePrecision.Int64 > 0 {
			return fmt.Sprintf("timestamp(%d) with time zone", c.DatetimePrecision.Int64)
		}
		return "timestamp with time zone"
	case "timestamp without time zone", "timestamp":
		if c.DatetimePrecision.Valid && c.DatetimePrecision.Int64 > 0 {
			return fmt.Sprintf("timestamp(%d) without time zone", c.DatetimePrecision.Int64)
		}
		return "timestamp without time zone"
	case "time with time zone", "timetz":
		if c.DatetimePrecision.Valid && c.DatetimePrecision.Int64 > 0 {
			return fmt.Sprintf("time(%d) with time zone", c.DatetimePrecision.Int64)
		}
		return "time with time zone"
	case "time without time zone", "time":
		if c.DatetimePrecision.Valid && c.DatetimePrecision.Int64 > 0 {
			return fmt.Sprintf("time(%d) without time zone", c.DatetimePrecision.Int64)
		}
		return "time without time zone"
	case "interval":
		if c.DatetimePrecision.Valid && c.DatetimePrecision.Int64 > 0 {
			return fmt.Sprintf("interval(%d)", c.DatetimePrecision.Int64)
		}
		return "interval"
	case "bit":
		if c.CharMaxLength.Valid && c.CharMaxLength.Int64 > 0 {
			return fmt.Sprintf("bit(%d)", c.CharMaxLength.Int64)
		}
		return "bit"
	case "bit varying", "varbit":
		if c.CharMaxLength.Valid && c.CharMaxLength.Int64 > 0 {
			return fmt.Sprintf("bit varying(%d)", c.CharMaxLength.Int64)
		}
		return "bit varying"
	default:
		return c.DataType
	}
}

func (s *postgresSchema) IndexSQL(tableName string) ([]string, error) {
	// pg_indexes.indexdef is the full CREATE INDEX statement.
	// Exclude indexes that back PRIMARY KEY / UNIQUE (they are already in CREATE TABLE).
	var indexdefs []string
	tx := s.db.Raw(`
		SELECT i.indexdef
		FROM pg_indexes i
		WHERE i.schemaname = current_schema() AND i.tablename = ?
		  AND NOT EXISTS (
		    SELECT 1 FROM pg_constraint c
		    WHERE c.conrelid = (current_schema()||'.'||?::text)::regclass
		      AND c.contype IN ('p','u') AND c.conname = i.indexname
		  )
		ORDER BY i.indexname
	`, tableName, tableName).Pluck("indexdef", &indexdefs)
	if tx.Error != nil {
		return nil, fmt.Errorf("list indexes for %s: %w", tableName, tx.Error)
	}
	out := make([]string, 0, len(indexdefs))
	for _, sql := range indexdefs {
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

// quoteIdent double-quotes identifier and escapes internal double quotes (PostgreSQL).
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
