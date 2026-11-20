package data

import (
	"context"
	"database/sql"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// postgresDialect 是 PostgreSQL 方言：双引号标识符、information_schema 与
// pg_catalog 内省、ALTER COLUMN 语法，以及 COMMENT ON 形式的注释。
type postgresDialect struct{}

// pgTypePattern 限定字段类型写法：类型名（可含空格，例如 timestamp with time zone），
// 可选的长度或精度括号，可选的数组后缀。类型名之后不允许出现引号、分号等符号。
var pgTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_ ]{0,48}(\(\s*\d{1,10}(\s*,\s*\d{1,10})?\s*\))?(\[\])?$`)

// pgDefaultPatterns 是允许的默认值写法。默认值会原样拼进 DDL，
// 因此只接受字面量和少数无副作用函数，其它表达式一律拒绝。
var pgDefaultPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?i:null|true|false)$`),
	regexp.MustCompile(`^-?\d+(\.\d+)?$`),
	regexp.MustCompile(`^'([^']|'')*'$`),
	regexp.MustCompile(`^(?i:now|current_timestamp|current_date|current_time|statement_timestamp|transaction_timestamp)\(\)$`),
	regexp.MustCompile(`^(?i:current_timestamp|current_date|current_time)$`),
	regexp.MustCompile(`^(?i:gen_random_uuid|uuid_generate_v4)\(\)$`),
}

func (postgresDialect) driverName() string {
	return "pgx"
}

// dsn 由连接串、用户名密码和目标库拼出 pgx 的 DSN。
// 用户名密码只放在 URL 的 userinfo 里，不拼字符串，避免特殊字符破坏解析。
func (postgresDialect) dsn(item bo.Datasource, database string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(item.Endpoint))
	if err != nil || parsed.Host == "" {
		return "", merr.Invalid("invalid datasource endpoint")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "postgres", "postgresql":
	default:
		return "", merr.Invalid("invalid datasource endpoint")
	}
	if database != "" {
		if !sqlNamePattern.MatchString(database) {
			return "", merr.Invalid("invalid database name")
		}
		parsed.Path = "/" + database
		parsed.RawPath = ""
	}
	if item.Username != "" {
		if item.Password != "" {
			parsed.User = url.UserPassword(item.Username, item.Password)
		} else {
			parsed.User = url.User(item.Username)
		}
	}
	query := parsed.Query()
	query.Set("application_name", "alioth")
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

// schemaName 空模式按 PostgreSQL 的默认模式 public 处理。
func (postgresDialect) schemaName(schema, _ string) string {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return "public"
	}
	return schema
}

// supportsDDLTransaction 返回 true：PostgreSQL 的 DDL 支持事务。
func (postgresDialect) supportsDDLTransaction() bool {
	return true
}

func (postgresDialect) quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (dialect postgresDialect) databases(ctx context.Context, db *sql.DB) ([]string, error) {
	return querySqlStrings(ctx, db,
		`SELECT datname FROM pg_database WHERE datistemplate = false AND has_database_privilege(datname, 'CONNECT') ORDER BY datname`)
}

// schemas 列出库里的用户模式。information_schema 和系统模式不展示。
func (postgresDialect) schemas(ctx context.Context, db *sql.DB, _ string) ([]string, error) {
	return querySqlStrings(ctx, db,
		`SELECT nspname FROM pg_namespace WHERE nspname <> 'information_schema' AND nspname NOT LIKE 'pg\_%' ORDER BY nspname`)
}

func (postgresDialect) tables(ctx context.Context, db *sql.DB, schema string) ([]bo.SqlTable, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = $1 AND table_type IN ('BASE TABLE', 'VIEW') ORDER BY table_name`,
		schema)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	tables := make([]bo.SqlTable, 0)
	for rows.Next() {
		var name, kind string
		if err := rows.Scan(&name, &kind); err != nil {
			return nil, sqlClientError(err)
		}
		if kind != "VIEW" {
			kind = "TABLE"
		}
		tables = append(tables, bo.SqlTable{Name: name, Type: kind})
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return tables, nil
}

func (postgresDialect) columns(ctx context.Context, db *sql.DB, schema, table string) ([]bo.SqlColumn, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT c.column_name, c.data_type, c.udt_name, (c.is_nullable = 'YES'),
		        COALESCE(c.column_default, ''), COALESCE(c.character_maximum_length, 0),
		        COALESCE(c.numeric_precision, 0), COALESCE(c.numeric_scale, 0),
		        c.ordinal_position,
		        EXISTS (
		          SELECT 1 FROM information_schema.table_constraints tc
		          JOIN information_schema.key_column_usage kcu
		            ON kcu.constraint_name = tc.constraint_name
		           AND kcu.table_schema = tc.table_schema
		           AND kcu.table_name = tc.table_name
		          WHERE tc.constraint_type = 'PRIMARY KEY'
		            AND tc.table_schema = c.table_schema
		            AND tc.table_name = c.table_name
		            AND kcu.column_name = c.column_name
		        ),
		        EXISTS (
		          SELECT 1 FROM information_schema.table_constraints tc
		          JOIN information_schema.key_column_usage kcu
		            ON kcu.constraint_name = tc.constraint_name
		           AND kcu.table_schema = tc.table_schema
		           AND kcu.table_name = tc.table_name
		          WHERE tc.constraint_type = 'UNIQUE'
		            AND tc.table_schema = c.table_schema
		            AND tc.table_name = c.table_name
		            AND kcu.column_name = c.column_name
		        )
		  FROM information_schema.columns c
		 WHERE c.table_schema = $1 AND c.table_name = $2
		 ORDER BY c.ordinal_position`,
		schema, table)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	columns := make([]bo.SqlColumn, 0)
	for rows.Next() {
		var column bo.SqlColumn
		if err := rows.Scan(&column.Name, &column.DataType, &column.UDTName, &column.Nullable,
			&column.DefaultValue, &column.Length, &column.NumericPrecision, &column.NumericScale,
			&column.Position, &column.PrimaryKey, &column.Unique); err != nil {
			return nil, sqlClientError(err)
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return columns, nil
}

// indexes 列出表的索引。索引字段按定义顺序返回；
// 表达式索引没有对应字段名，只保留定义。
func (postgresDialect) indexes(ctx context.Context, db *sql.DB, schema, table string) ([]bo.SqlIndex, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT i.relname, ix.indisunique, ix.indisprimary, pg_get_indexdef(ix.indexrelid), COALESCE(a.attname, '')
		   FROM pg_index ix
		   JOIN pg_class i ON i.oid = ix.indexrelid
		   JOIN pg_class t ON t.oid = ix.indrelid
		   JOIN pg_namespace n ON n.oid = t.relnamespace
		   LEFT JOIN LATERAL unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord) ON true
		   LEFT JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = k.attnum
		  WHERE n.nspname = $1 AND t.relname = $2
		  ORDER BY i.relname, k.ord`,
		schema, table)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	indexes := make([]bo.SqlIndex, 0)
	positions := make(map[string]int, 8)
	for rows.Next() {
		var name, definition, column string
		var unique, primary bool
		if err := rows.Scan(&name, &unique, &primary, &definition, &column); err != nil {
			return nil, sqlClientError(err)
		}
		position, ok := positions[name]
		if !ok {
			position = len(indexes)
			positions[name] = position
			indexes = append(indexes, bo.SqlIndex{Name: name, Unique: unique, Primary: primary, Definition: definition})
		}
		if column != "" {
			indexes[position].Columns = append(indexes[position].Columns, column)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return indexes, nil
}

func (postgresDialect) primaryKeys(ctx context.Context, db *sql.DB, schema, table string) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT a.attname
		   FROM pg_index i
		   JOIN pg_class c ON c.oid = i.indrelid
		   JOIN pg_namespace n ON n.oid = c.relnamespace
		   JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY (i.indkey)
		  WHERE n.nspname = $1 AND c.relname = $2 AND i.indisprimary
		  ORDER BY array_position(i.indkey, a.attnum)`, schema, table)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	keys := make([]string, 0, 2)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, sqlClientError(err)
		}
		keys = append(keys, name)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return keys, nil
}

// tableDataStatement 生成分页读取表数据的语句。有主键按主键排序，
// 没有主键按物理行号 ctid 排序，让翻页在同一时刻的物理顺序上保持稳定。
func (dialect postgresDialect) tableDataStatement(schema, table string, primaryKeys []string, limit, offset int) (string, error) {
	order := "ctid"
	if len(primaryKeys) > 0 {
		quoted := make([]string, 0, len(primaryKeys))
		for _, column := range primaryKeys {
			quoted = append(quoted, dialect.quoteIdentifier(column))
		}
		order = strings.Join(quoted, ", ")
	}
	return "SELECT * FROM " + pgTableName(dialect, schema, table) + " ORDER BY " + order +
		" LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset), nil
}

// createStatements 生成建表语句。主键字段强制非空，字段名不允许重复。
func (dialect postgresDialect) createStatements(_ context.Context, _ *sql.DB, in bo.CreateSqlTableInput, schema string) ([]string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return nil, err
	}
	if len(in.Columns) == 0 {
		return nil, merr.Invalid("columns are required")
	}
	definitions, err := dialect.columnDefinitions(in.Columns)
	if err != nil {
		return nil, err
	}
	primaryKeys := make([]string, 0, 1)
	statement := "CREATE TABLE "
	if in.IfNotExists {
		statement += "IF NOT EXISTS "
	}
	target := pgTableName(dialect, schema, in.Table)
	statement += target + " (\n  " + strings.Join(definitions, ",\n  ")
	for _, column := range in.Columns {
		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, dialect.quoteIdentifier(column.Name))
		}
	}
	if len(primaryKeys) > 0 {
		statement += ",\n  PRIMARY KEY (" + strings.Join(primaryKeys, ", ") + ")"
	}
	statement += "\n)"
	statements := []string{statement}
	comment, err := validateSqlComment(in.Comment)
	if err != nil {
		return nil, err
	}
	if comment != "" {
		statements = append(statements, "COMMENT ON TABLE "+target+" IS "+quoteSqlLiteral(comment))
	}
	for _, column := range in.Columns {
		columnComment, err := validateSqlComment(column.Comment)
		if err != nil {
			return nil, err
		}
		if columnComment == "" {
			continue
		}
		statements = append(statements, "COMMENT ON COLUMN "+target+"."+dialect.quoteIdentifier(column.Name)+" IS "+quoteSqlLiteral(columnComment))
	}
	return statements, nil
}

// columnDefinitions 生成字段定义，每个字段名只允许出现一次。
func (dialect postgresDialect) columnDefinitions(columns []bo.SqlColumnDefinition) ([]string, error) {
	seen := make(map[string]struct{}, len(columns))
	definitions := make([]string, 0, len(columns))
	for _, column := range columns {
		name := strings.TrimSpace(column.Name)
		if err := requireSqlName(name); err != nil {
			return nil, err
		}
		if _, ok := seen[name]; ok {
			return nil, merr.Invalid("duplicate sql column")
		}
		seen[name] = struct{}{}
		columnType, err := dialect.validateType(column.Type)
		if err != nil {
			return nil, err
		}
		defaultValue, err := dialect.validateDefault(column.DefaultValue)
		if err != nil {
			return nil, err
		}
		definition := dialect.quoteIdentifier(name) + " " + columnType
		if !column.Nullable || column.PrimaryKey {
			definition += " NOT NULL"
		}
		if defaultValue != "" {
			definition += " DEFAULT " + defaultValue
		}
		if column.Unique && !column.PrimaryKey {
			definition += " UNIQUE"
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

// alterStatements 生成改表语句。至少要有一次修改，字段名不允许重复。
// 改名放在最后，避免后续语句指向已经被改名的表。
func (dialect postgresDialect) alterStatements(_ context.Context, _ *sql.DB, in bo.AlterSqlTableInput, schema string) ([]string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return nil, err
	}
	target := pgTableName(dialect, schema, in.Table)
	statements := make([]string, 0, 8)
	if len(in.AddColumns) > 0 {
		definitions, err := dialect.columnDefinitions(in.AddColumns)
		if err != nil {
			return nil, err
		}
		for _, definition := range definitions {
			statements = append(statements, "ALTER TABLE "+target+" ADD COLUMN "+definition)
		}
	}
	dropped := make(map[string]struct{}, len(in.DropColumns))
	for _, column := range in.DropColumns {
		name := strings.TrimSpace(column)
		if err := requireSqlName(name); err != nil {
			return nil, err
		}
		if _, ok := dropped[name]; ok {
			return nil, merr.Invalid("duplicate sql column")
		}
		dropped[name] = struct{}{}
		statements = append(statements, "ALTER TABLE "+target+" DROP COLUMN "+dialect.quoteIdentifier(name))
	}
	altered := make(map[string]struct{}, len(in.AlterColumns))
	for _, column := range in.AlterColumns {
		name := strings.TrimSpace(column.Name)
		if err := requireSqlName(name); err != nil {
			return nil, err
		}
		if _, ok := altered[name]; ok {
			return nil, merr.Invalid("duplicate sql column")
		}
		altered[name] = struct{}{}
		quoted := dialect.quoteIdentifier(name)
		changes := 0
		if column.Type != "" {
			columnType, err := dialect.validateType(column.Type)
			if err != nil {
				return nil, err
			}
			statements = append(statements, "ALTER TABLE "+target+" ALTER COLUMN "+quoted+" TYPE "+columnType)
			changes++
		}
		if column.Nullable != nil {
			if *column.Nullable {
				statements = append(statements, "ALTER TABLE "+target+" ALTER COLUMN "+quoted+" DROP NOT NULL")
			} else {
				statements = append(statements, "ALTER TABLE "+target+" ALTER COLUMN "+quoted+" SET NOT NULL")
			}
			changes++
		}
		if column.DropDefault {
			statements = append(statements, "ALTER TABLE "+target+" ALTER COLUMN "+quoted+" DROP DEFAULT")
			changes++
		} else if column.DefaultValue != "" {
			defaultValue, err := dialect.validateDefault(column.DefaultValue)
			if err != nil {
				return nil, err
			}
			statements = append(statements, "ALTER TABLE "+target+" ALTER COLUMN "+quoted+" SET DEFAULT "+defaultValue)
			changes++
		}
		if changes == 0 {
			return nil, merr.Invalid("nothing to alter")
		}
	}
	if in.RenameTo != "" {
		if err := requireSqlName(in.RenameTo); err != nil {
			return nil, err
		}
		if in.RenameTo == in.Table {
			return nil, merr.Invalid("invalid sql rename")
		}
		statements = append(statements, "ALTER TABLE "+target+" RENAME TO "+dialect.quoteIdentifier(in.RenameTo))
	}
	if len(statements) == 0 {
		return nil, merr.Invalid("nothing to alter")
	}
	return statements, nil
}

func (dialect postgresDialect) dropStatement(in bo.DropSqlTableInput, schema string) (string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return "", err
	}
	statement := "DROP TABLE "
	if in.IfExists {
		statement += "IF EXISTS "
	}
	statement += pgTableName(dialect, schema, in.Table)
	return statement, nil
}

// validateType 校验字段类型写法，返回去掉多余空白后的类型。
func (postgresDialect) validateType(raw string) (string, error) {
	value := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if !pgTypePattern.MatchString(value) {
		return "", merr.Invalid("invalid sql column type")
	}
	return strings.ToLower(value), nil
}

// validateDefault 校验默认值写法。留空表示没有默认值。
func (postgresDialect) validateDefault(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	for _, pattern := range pgDefaultPatterns {
		if pattern.MatchString(value) {
			return value, nil
		}
	}
	return "", merr.Invalid("invalid sql default value")
}

// pgTableName 返回带模式的表名，标识符都加了双引号。
func pgTableName(dialect postgresDialect, schema, table string) string {
	return dialect.quoteIdentifier(schema) + "." + dialect.quoteIdentifier(table)
}

// querySqlStrings 执行一条只返回一列文本的查询。两种方言共用。
func querySqlStrings(ctx context.Context, db *sql.DB, statement string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	values := make([]string, 0, 16)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, sqlClientError(err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return values, nil
}
