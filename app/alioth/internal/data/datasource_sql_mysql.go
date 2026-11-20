package data

import (
	"context"
	"database/sql"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// mysqlDialect 是 MySQL 方言：反引号标识符、information_schema 内省、
// MODIFY COLUMN 形式的字段修改，以及内联在字段定义里的 COMMENT。
// MySQL 里模式和库是同一层，schema 为空时用当前连接的库兜底。
type mysqlDialect struct{}

// mysqlTypePattern 限定字段类型写法：类型名、可选的长度或精度括号，
// 以及可选的 unsigned / zerofill 修饰。其它符号一律拒绝。
var mysqlTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_ ]{0,48}(\(\s*\d{1,10}(\s*,\s*\d{1,10})?\s*\))?(\s+(?i:unsigned|signed))?(\s+(?i:zerofill))?$`)

// mysqlDefaultPatterns 是允许的默认值写法。默认值会原样拼进 DDL，
// 因此只接受字面量和 CURRENT_TIMESTAMP 一类无副作用表达式。
var mysqlDefaultPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?i:null|true|false)$`),
	regexp.MustCompile(`^-?\d+(\.\d+)?$`),
	regexp.MustCompile(`^'([^']|'')*'$`),
	regexp.MustCompile(`^(?i:current_timestamp|current_date|current_time)(\(\d\))?$`),
	regexp.MustCompile(`^(?i:now|uuid)\(\)$`),
}

// mysqlColumnState 是改表时读到的现有字段定义。MODIFY COLUMN 会整体替换
// 字段定义，所以没改的部分要原样补回去。
type mysqlColumnState struct {
	columnType   string
	charset      string
	collation    string
	defaultValue string
	extra        string
	comment      string
	nullable     bool
}

func (mysqlDialect) driverName() string {
	return "mysql"
}

// dsn 由 mysql:// 连接串、用户名密码和目标库生成 go-sql-driver 的连接串。
// 用户名密码优先取数据源字段，连接串里的 userinfo 作为兜底。
func (mysqlDialect) dsn(item bo.Datasource, database string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(item.Endpoint))
	if err != nil || parsed.Host == "" {
		return "", merr.Invalid("invalid datasource endpoint")
	}
	if strings.ToLower(parsed.Scheme) != "mysql" {
		return "", merr.Invalid("invalid datasource endpoint")
	}
	user, password := "", ""
	if parsed.User != nil {
		user = parsed.User.Username()
		password, _ = parsed.User.Password()
	}
	if item.Username != "" {
		user = item.Username
		password = item.Password
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if database != "" {
		if !sqlNamePattern.MatchString(database) {
			return "", merr.Invalid("invalid database name")
		}
		dbName = database
	}
	if dbName == "" {
		return "", merr.Invalid("database is required")
	}
	config := mysql.NewConfig()
	config.User = user
	config.Passwd = password
	config.Net = "tcp"
	config.Addr = parsed.Host
	config.DBName = dbName
	config.ParseTime = true
	config.Loc = time.Local
	config.Timeout = sqlConnectTimeout
	config.ReadTimeout = sqlQueryTimeout
	config.WriteTimeout = sqlQueryTimeout
	config.Params = map[string]string{"charset": "utf8mb4"}
	return config.FormatDSN(), nil
}

// schemaName 在 MySQL 里模式和库同名，空值交给内省语句用当前库兜底。
func (mysqlDialect) schemaName(schema, database string) string {
	if schema = strings.TrimSpace(schema); schema != "" {
		return schema
	}
	return strings.TrimSpace(database)
}

func (mysqlDialect) quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// supportsDDLTransaction 返回 false：MySQL 的 DDL 会隐式提交，
// 表结构变更不能靠事务回滚，按语句顺序执行。
func (mysqlDialect) supportsDDLTransaction() bool {
	return false
}

func (mysqlDialect) databases(ctx context.Context, db *sql.DB) ([]string, error) {
	return querySqlStrings(ctx, db,
		`SELECT SCHEMA_NAME FROM information_schema.SCHEMATA ORDER BY SCHEMA_NAME`)
}

func (mysqlDialect) schemas(ctx context.Context, db *sql.DB, _ string) ([]string, error) {
	return querySqlStrings(ctx, db,
		`SELECT SCHEMA_NAME FROM information_schema.SCHEMATA
		  WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
		  ORDER BY SCHEMA_NAME`)
}

func (dialect mysqlDialect) tables(ctx context.Context, db *sql.DB, schema string) ([]bo.SqlTable, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES
		  WHERE TABLE_SCHEMA = COALESCE(NULLIF(?, ''), DATABASE())
		    AND TABLE_TYPE IN ('BASE TABLE', 'VIEW')
		  ORDER BY TABLE_NAME`, schema)
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
		if strings.EqualFold(kind, "VIEW") {
			kind = "VIEW"
		} else {
			kind = "TABLE"
		}
		tables = append(tables, bo.SqlTable{Name: name, Type: kind})
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return tables, nil
}

func (dialect mysqlDialect) columns(ctx context.Context, db *sql.DB, schema, table string) ([]bo.SqlColumn, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT COLUMN_NAME, COLUMN_TYPE, DATA_TYPE, (IS_NULLABLE = 'YES'),
		        COALESCE(COLUMN_DEFAULT, ''), COALESCE(CHARACTER_MAXIMUM_LENGTH, 0),
		        COALESCE(NUMERIC_PRECISION, 0), COALESCE(NUMERIC_SCALE, 0),
		        ORDINAL_POSITION, (COLUMN_KEY = 'PRI'), (COLUMN_KEY = 'UNI')
		   FROM information_schema.COLUMNS
		  WHERE TABLE_SCHEMA = COALESCE(NULLIF(?, ''), DATABASE()) AND TABLE_NAME = ?
		  ORDER BY ORDINAL_POSITION`, schema, table)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	columns := make([]bo.SqlColumn, 0)
	for rows.Next() {
		var column bo.SqlColumn
		var nullable, primaryKey, unique int
		if err := rows.Scan(&column.Name, &column.DataType, &column.UDTName, &nullable,
			&column.DefaultValue, &column.Length, &column.NumericPrecision, &column.NumericScale,
			&column.Position, &primaryKey, &unique); err != nil {
			return nil, sqlClientError(err)
		}
		column.Nullable = nullable != 0
		column.PrimaryKey = primaryKey != 0
		column.Unique = unique != 0
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return columns, nil
}

// indexes 按索引名聚合字段，定义按类型拼出来。表达式索引的字段名为空。
func (dialect mysqlDialect) indexes(ctx context.Context, db *sql.DB, schema, table string) ([]bo.SqlIndex, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT INDEX_NAME, COALESCE(COLUMN_NAME, ''), NON_UNIQUE
		   FROM information_schema.STATISTICS
		  WHERE TABLE_SCHEMA = COALESCE(NULLIF(?, ''), DATABASE()) AND TABLE_NAME = ?
		  ORDER BY INDEX_NAME, SEQ_IN_INDEX`, schema, table)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	indexes := make([]bo.SqlIndex, 0)
	positions := make(map[string]int, 8)
	for rows.Next() {
		var name, column string
		var nonUnique int
		if err := rows.Scan(&name, &column, &nonUnique); err != nil {
			return nil, sqlClientError(err)
		}
		position, ok := positions[name]
		if !ok {
			position = len(indexes)
			positions[name] = position
			indexes = append(indexes, bo.SqlIndex{Name: name, Unique: nonUnique == 0, Primary: strings.EqualFold(name, "PRIMARY")})
		}
		if column != "" {
			indexes[position].Columns = append(indexes[position].Columns, column)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	for index := range indexes {
		indexes[index].Definition = mysqlIndexDefinition(dialect, indexes[index])
	}
	return indexes, nil
}

// mysqlIndexDefinition 把索引拼成可读的定义文本。
func mysqlIndexDefinition(dialect mysqlDialect, index bo.SqlIndex) string {
	columns := make([]string, 0, len(index.Columns))
	for _, column := range index.Columns {
		columns = append(columns, dialect.quoteIdentifier(column))
	}
	joined := "(" + strings.Join(columns, ", ") + ")"
	switch {
	case index.Primary:
		return "PRIMARY KEY " + joined
	case index.Unique:
		return "UNIQUE INDEX " + dialect.quoteIdentifier(index.Name) + " " + joined
	default:
		return "INDEX " + dialect.quoteIdentifier(index.Name) + " " + joined
	}
}

func (mysqlDialect) primaryKeys(ctx context.Context, db *sql.DB, schema, table string) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT COLUMN_NAME FROM information_schema.KEY_COLUMN_USAGE
		  WHERE TABLE_SCHEMA = COALESCE(NULLIF(?, ''), DATABASE())
		    AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'
		  ORDER BY ORDINAL_POSITION`, schema, table)
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

// tableDataStatement 生成分页读取表数据的语句。有主键按主键排序；
// 没有主键时使用存储顺序，翻页顺序由 MySQL 决定。
func (dialect mysqlDialect) tableDataStatement(schema, table string, primaryKeys []string, limit, offset int) (string, error) {
	order := ""
	if len(primaryKeys) > 0 {
		quoted := make([]string, 0, len(primaryKeys))
		for _, column := range primaryKeys {
			quoted = append(quoted, dialect.quoteIdentifier(column))
		}
		order = " ORDER BY " + strings.Join(quoted, ", ")
	}
	return "SELECT * FROM " + mysqlTableName(dialect, schema, table) + order +
		" LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset), nil
}

// createStatements 生成建表语句。主键字段强制非空，字段名不允许重复，
// 注释内联在字段定义和表定义里。
func (dialect mysqlDialect) createStatements(_ context.Context, _ *sql.DB, in bo.CreateSqlTableInput, schema string) ([]string, error) {
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
	statement += mysqlTableName(dialect, schema, in.Table) + " (\n  " + strings.Join(definitions, ",\n  ")
	for _, column := range in.Columns {
		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, dialect.quoteIdentifier(column.Name))
		}
	}
	if len(primaryKeys) > 0 {
		statement += ",\n  PRIMARY KEY (" + strings.Join(primaryKeys, ", ") + ")"
	}
	statement += "\n)"
	comment, err := validateSqlComment(in.Comment)
	if err != nil {
		return nil, err
	}
	if comment != "" {
		statement += " COMMENT=" + quoteSqlLiteral(comment)
	}
	return []string{statement}, nil
}

// columnDefinitions 生成字段定义，每个字段名只允许出现一次。
func (dialect mysqlDialect) columnDefinitions(columns []bo.SqlColumnDefinition) ([]string, error) {
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
		comment, err := validateSqlComment(column.Comment)
		if err != nil {
			return nil, err
		}
		if comment != "" {
			definition += " COMMENT " + quoteSqlLiteral(comment)
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

// alterStatements 生成改表语句。改名放在最后，避免后续语句指向已经改名的表。
func (dialect mysqlDialect) alterStatements(ctx context.Context, db *sql.DB, in bo.AlterSqlTableInput, schema string) ([]string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return nil, err
	}
	target := mysqlTableName(dialect, schema, in.Table)
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
	if len(in.AlterColumns) > 0 {
		states, err := loadMysqlColumnStates(ctx, db, schema, in.Table)
		if err != nil {
			return nil, err
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
			state, ok := states[name]
			if !ok {
				return nil, merr.Invalid("unknown sql column")
			}
			definition, err := dialect.modifyDefinition(column, state)
			if err != nil {
				return nil, err
			}
			statements = append(statements, "ALTER TABLE "+target+" MODIFY COLUMN "+definition)
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

// modifyDefinition 把一次字段修改合并到现有定义上。MODIFY COLUMN 会整体
// 替换字段，所以没改的类型、字符集、默认值、自增和注释都要原样补回去。
func (dialect mysqlDialect) modifyDefinition(column bo.SqlAlterColumn, state mysqlColumnState) (string, error) {
	columnType := state.columnType
	if column.Type != "" {
		validated, err := dialect.validateType(column.Type)
		if err != nil {
			return "", err
		}
		columnType = validated
	}
	if columnType == "" {
		return "", merr.Invalid("invalid sql column type")
	}
	nullable := state.nullable
	if column.Nullable != nil {
		nullable = *column.Nullable
	}
	var defaultValue string
	switch {
	case column.DropDefault:
		// 不写 DEFAULT 就是清除默认值。
	case column.DefaultValue != "":
		validated, err := dialect.validateDefault(column.DefaultValue)
		if err != nil {
			return "", err
		}
		defaultValue = validated
	case strings.Contains(strings.ToLower(state.extra), "default_generated"):
		// 表达式默认值（例如 CURRENT_TIMESTAMP）原样保留。
		defaultValue = state.defaultValue
	case state.defaultValue != "":
		defaultValue = quoteSqlLiteral(state.defaultValue)
	}
	definition := dialect.quoteIdentifier(column.Name) + " " + columnType
	if state.charset != "" && state.collation != "" {
		definition += " CHARACTER SET " + state.charset + " COLLATE " + state.collation
	}
	if nullable {
		definition += " NULL"
	} else {
		definition += " NOT NULL"
	}
	if defaultValue != "" {
		definition += " DEFAULT " + defaultValue
	}
	if strings.Contains(strings.ToLower(state.extra), "auto_increment") {
		definition += " AUTO_INCREMENT"
	}
	if state.comment != "" {
		definition += " COMMENT " + quoteSqlLiteral(state.comment)
	}
	return definition, nil
}

// loadMysqlColumnStates 读取表现有的字段定义，供 MODIFY COLUMN 合并。
func loadMysqlColumnStates(ctx context.Context, db *sql.DB, schema, table string) (map[string]mysqlColumnState, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COALESCE(COLUMN_DEFAULT, ''),
		        COALESCE(EXTRA, ''), COALESCE(COLUMN_COMMENT, ''),
		        COALESCE(CHARACTER_SET_NAME, ''), COALESCE(COLLATION_NAME, '')
		   FROM information_schema.COLUMNS
		  WHERE TABLE_SCHEMA = COALESCE(NULLIF(?, ''), DATABASE()) AND TABLE_NAME = ?
		  ORDER BY ORDINAL_POSITION`, schema, table)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	states := make(map[string]mysqlColumnState, 8)
	for rows.Next() {
		var name, nullable string
		var state mysqlColumnState
		if err := rows.Scan(&name, &state.columnType, &nullable, &state.defaultValue,
			&state.extra, &state.comment, &state.charset, &state.collation); err != nil {
			return nil, sqlClientError(err)
		}
		state.nullable = strings.EqualFold(nullable, "YES")
		states[name] = state
	}
	if err := rows.Err(); err != nil {
		return nil, sqlClientError(err)
	}
	return states, nil
}

func (dialect mysqlDialect) dropStatement(in bo.DropSqlTableInput, schema string) (string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return "", err
	}
	statement := "DROP TABLE "
	if in.IfExists {
		statement += "IF EXISTS "
	}
	statement += mysqlTableName(dialect, schema, in.Table)
	return statement, nil
}

// validateType 校验字段类型写法，返回去掉多余空白后的类型。
func (mysqlDialect) validateType(raw string) (string, error) {
	value := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if !mysqlTypePattern.MatchString(value) {
		return "", merr.Invalid("invalid sql column type")
	}
	return strings.ToLower(value), nil
}

// validateDefault 校验默认值写法。留空表示没有默认值。
func (mysqlDialect) validateDefault(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	for _, pattern := range mysqlDefaultPatterns {
		if pattern.MatchString(value) {
			return value, nil
		}
	}
	return "", merr.Invalid("invalid sql default value")
}

// mysqlTableName 返回表名。MySQL 里 schema 就是库，为空时只写表名，
// 由当前连接选择库。
func mysqlTableName(dialect mysqlDialect, schema, table string) string {
	if schema == "" {
		return dialect.quoteIdentifier(table)
	}
	return dialect.quoteIdentifier(schema) + "." + dialect.quoteIdentifier(table)
}
