package data

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasourcesqlaudit"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// sqlConnectTimeout 是连接用户数据库的超时。地址填错时尽快报错，
	// 不让页面一直挂着。
	sqlConnectTimeout = 10 * time.Second
	// sqlQueryTimeout 是一次查询或结构变更的上限。
	sqlQueryTimeout = 30 * time.Second
	// sqlAuditTimeout 是写审计的独立超时。执行失败也要留下审计，
	// 不能因为客户端断开就连审计一起丢。
	sqlAuditTimeout = 5 * time.Second
	// sqlDefaultRows 是查询未指定行数时的默认上限，sqlMaxRows 是绝对上限。
	sqlDefaultRows = 100
	sqlMaxRows     = 1000
	// sqlMaxAuditError 是审计里保存的失败原因长度。
	sqlMaxAuditError = 300
	// sqlMaxTablePageRows 是表数据分页的每页上限。
	sqlMaxTablePageRows = 500
)

// sqlNamePattern 限定库名、模式名、表名和字段名。只接受 PostgreSQL
// 不加引号也安全的标识符，避免把用户输入拼进 DDL 时出现注入。
var sqlNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

// sqlTypePattern 限定字段类型写法：类型名（可含空格，例如 timestamp with time zone），
// 可选的长度或精度括号，可选的数组后缀。类型名之后不允许出现引号、分号等符号。
var sqlTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_ ]{0,48}(\(\s*\d{1,10}(\s*,\s*\d{1,10})?\s*\))?(\[\])?$`)

// sqlDefaultPatterns 是允许的默认值写法。默认值会原样拼进 DDL，
// 因此只接受字面量和少数无副作用函数，其它表达式一律拒绝。
var sqlDefaultPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?i:null|true|false)$`),
	regexp.MustCompile(`^-?\d+(\.\d+)?$`),
	regexp.MustCompile(`^'([^']|'')*'$`),
	regexp.MustCompile(`^(?i:now|current_timestamp|current_date|current_time|statement_timestamp|transaction_timestamp)\(\)$`),
	regexp.MustCompile(`^(?i:current_timestamp|current_date|current_time)$`),
	regexp.MustCompile(`^(?i:gen_random_uuid|uuid_generate_v4)\(\)$`),
}

// sqlReadOnlyKeywords 是只读查询允许的首个关键字。
var sqlReadOnlyKeywords = map[string]struct{}{
	"SELECT":  {},
	"WITH":    {},
	"SHOW":    {},
	"EXPLAIN": {},
	"VALUES":  {},
	"TABLE":   {},
}

// sqlDataSourceFor 打开用户数据库的连接。每次请求只开一个连接，
// 用完即关，不在进程里缓存用户数据库的会话。
func (r *datasourceRepo) sqlDataSourceFor(ctx context.Context, item bo.Datasource, database string) (*sql.DB, error) {
	dsn, err := sqlDSN(item, database)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, sqlClientError(err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	pingCtx, cancel := context.WithTimeout(ctx, sqlConnectTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, sqlClientError(err)
	}
	return db, nil
}

// sqlDSN 由连接串、用户名密码和目标库拼出 pgx 的 DSN。
// 用户名密码只放在 URL 的 userinfo 里，不拼字符串，避免特殊字符破坏解析。
func sqlDSN(item bo.Datasource, database string) (string, error) {
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

// sqlClientError 把数据库错误压成一行交给调用方。连接失败、权限不足和语句错误
// 都是用户可修正的配置问题，按客户端错误返回；摘要长度受限，避免整段堆栈进页面。
func sqlClientError(err error) error {
	reason := strings.Join(strings.Fields(err.Error()), " ")
	if reason == "" {
		reason = "sql statement failed"
	}
	runes := []rune(reason)
	if len(runes) > maxUpstreamReason {
		reason = string(runes[:maxUpstreamReason]) + "…"
	}
	return merr.Client("sql statement failed: " + reason)
}

// SqlDatabases 列出当前连接可以访问的库。
func (r *datasourceRepo) SqlDatabases(ctx context.Context, item bo.Datasource) ([]string, error) {
	return r.querySqlStrings(ctx, item, "",
		`SELECT datname FROM pg_database WHERE datistemplate = false AND has_database_privilege(datname, 'CONNECT') ORDER BY datname`)
}

// SqlSchemas 列出库里的用户模式。information_schema 和系统模式不展示。
func (r *datasourceRepo) SqlSchemas(ctx context.Context, item bo.Datasource, database string) ([]string, error) {
	return r.querySqlStrings(ctx, item, database,
		`SELECT nspname FROM pg_namespace WHERE nspname <> 'information_schema' AND nspname NOT LIKE 'pg\_%' ORDER BY nspname`)
}

// SqlTables 列出模式里的表和视图。
func (r *datasourceRepo) SqlTables(ctx context.Context, item bo.Datasource, database, schema string) ([]bo.SqlTable, error) {
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	rows, err := db.QueryContext(queryCtx,
		`SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = $1 AND table_type IN ('BASE TABLE', 'VIEW') ORDER BY table_name`,
		sqlSchema(schema))
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

// SqlTableColumns 列出表的字段、默认值、主键和唯一约束。
func (r *datasourceRepo) SqlTableColumns(ctx context.Context, item bo.Datasource, database, schema, table string) ([]bo.SqlColumn, error) {
	if err := requireSqlName(table); err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	rows, err := db.QueryContext(queryCtx,
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
		sqlSchema(schema), table)
	if err != nil {
		return nil, err
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

// SqlTableIndexes 列出表的索引。索引字段按定义顺序返回；
// 表达式索引没有对应字段名，只保留定义。
func (r *datasourceRepo) SqlTableIndexes(ctx context.Context, item bo.Datasource, database, schema, table string) ([]bo.SqlIndex, error) {
	if err := requireSqlName(table); err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	rows, err := db.QueryContext(queryCtx,
		`SELECT i.relname, ix.indisunique, ix.indisprimary, pg_get_indexdef(ix.indexrelid), COALESCE(a.attname, '')
		   FROM pg_index ix
		   JOIN pg_class i ON i.oid = ix.indexrelid
		   JOIN pg_class t ON t.oid = ix.indrelid
		   JOIN pg_namespace n ON n.oid = t.relnamespace
		   LEFT JOIN LATERAL unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord) ON true
		   LEFT JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = k.attnum
		  WHERE n.nspname = $1 AND t.relname = $2
		  ORDER BY i.relname, k.ord`,
		sqlSchema(schema), table)
	if err != nil {
		return nil, err
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

// QuerySql 在只读事务里执行一条查询。只允许读取类语句，行数按上限截断。
func (r *datasourceRepo) QuerySql(ctx context.Context, item bo.Datasource, database, statement string, maxRows int) (*bo.SqlQueryResult, error) {
	if err := requireReadOnlySql(statement); err != nil {
		return nil, err
	}
	if maxRows <= 0 {
		maxRows = sqlDefaultRows
	}
	if maxRows > sqlMaxRows {
		maxRows = sqlMaxRows
	}
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	// 只读事务兜底：即使语句绕过了关键字检查，也不能写库。
	tx, err := db.BeginTx(queryCtx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	started := time.Now()
	rows, err := tx.QueryContext(queryCtx, statement)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	names, values, truncated, err := readSqlRows(rows, maxRows)
	if err != nil {
		return nil, err
	}
	return &bo.SqlQueryResult{
		Columns: names, Rows: values, Truncated: truncated,
		ElapsedMS: time.Since(started).Milliseconds(),
	}, nil
}

// readSqlRows 把结果集按列顺序读成字符串单元格。超过 maxRows 时停在 maxRows 并标记截断。
func readSqlRows(rows *sql.Rows, maxRows int) ([]string, [][]string, bool, error) {
	names, err := rows.Columns()
	if err != nil {
		return nil, nil, false, sqlClientError(err)
	}
	values := make([][]string, 0, 16)
	cells := make([]any, len(names))
	for index := range cells {
		var value any
		cells[index] = &value
	}
	truncated := false
	for rows.Next() {
		if len(values) >= maxRows {
			truncated = true
			break
		}
		if err := rows.Scan(cells...); err != nil {
			return nil, nil, false, sqlClientError(err)
		}
		row := make([]string, len(names))
		for index, cell := range cells {
			row[index] = sqlCellText(*(cell.(*any)))
		}
		values = append(values, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, sqlClientError(err)
	}
	return names, values, truncated, nil
}

// SqlTableData 分页读取表数据。排序用主键，没有主键时用物理行号 ctid，
// 让翻页在同一时刻的物理顺序上保持稳定。
func (r *datasourceRepo) SqlTableData(ctx context.Context, item bo.Datasource, database, schema, table string, page, pageSize int) (*bo.SqlTableRows, error) {
	if err := requireSqlName(table); err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = sqlDefaultRows
	}
	if pageSize > sqlMaxTablePageRows {
		pageSize = sqlMaxTablePageRows
	}
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// 多取一行用来判断是否还有下一页。
	statement := buildSqlTableDataStatement(schema, table, nil, pageSize+1, (page-1)*pageSize)
	primaryKeys, err := r.sqlPrimaryKeys(ctx, db, sqlSchema(schema), table)
	if err != nil {
		return nil, err
	}
	if len(primaryKeys) > 0 {
		statement = buildSqlTableDataStatement(schema, table, primaryKeys, pageSize+1, (page-1)*pageSize)
	}
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	started := time.Now()
	rows, err := db.QueryContext(queryCtx, statement)
	if err != nil {
		return nil, sqlClientError(err)
	}
	defer rows.Close()
	names, values, _, err := readSqlRows(rows, pageSize+1)
	if err != nil {
		return nil, err
	}
	result := &bo.SqlTableRows{Columns: names}
	if len(values) > pageSize {
		result.HasMore = true
		values = values[:pageSize]
	}
	result.Rows = values
	result.ElapsedMS = time.Since(started).Milliseconds()
	return result, nil
}

// sqlPrimaryKeys 读取表的主键字段，按主键定义顺序返回。
func (r *datasourceRepo) sqlPrimaryKeys(ctx context.Context, db *sql.DB, schema, table string) ([]string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	rows, err := db.QueryContext(queryCtx,
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

// buildSqlTableDataStatement 生成分页读取表数据的语句。表名和字段名都加引号，
// limit 和 offset 由调用方计算，不拼接任何用户输入。
func buildSqlTableDataStatement(schema, table string, primaryKeys []string, limit, offset int) string {
	order := "ctid"
	if len(primaryKeys) > 0 {
		quoted := make([]string, 0, len(primaryKeys))
		for _, column := range primaryKeys {
			quoted = append(quoted, quoteSqlIdentifier(column))
		}
		order = strings.Join(quoted, ", ")
	}
	return "SELECT * FROM " + sqlTableName(schema, table) + " ORDER BY " + order +
		" LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
}

// CreateSqlTable 建表并写审计。表注释和字段注释作为同一次事务里的后续语句。
func (r *datasourceRepo) CreateSqlTable(ctx context.Context, item bo.Datasource, in bo.CreateSqlTableInput, operatorUID string) (string, error) {
	statements, err := buildCreateSqlTable(in)
	if err != nil {
		return "", err
	}
	if err := r.execSqlDDL(ctx, item, in.Database, datasourcesqlaudit.ActionCREATE_TABLE, sqlSchema(in.Schema), in.Table, statements, operatorUID); err != nil {
		return "", err
	}
	return strings.Join(statements, ";\n"), nil
}

// AlterSqlTable 改表并写审计。rename、加字段、删字段和改字段可以一次提交。
func (r *datasourceRepo) AlterSqlTable(ctx context.Context, item bo.Datasource, in bo.AlterSqlTableInput, operatorUID string) ([]string, error) {
	statements, err := buildAlterSqlTable(in)
	if err != nil {
		return nil, err
	}
	if err := r.execSqlDDL(ctx, item, in.Database, datasourcesqlaudit.ActionALTER_TABLE, sqlSchema(in.Schema), in.Table, statements, operatorUID); err != nil {
		return nil, err
	}
	return statements, nil
}

// DropSqlTable 删表并写审计。
func (r *datasourceRepo) DropSqlTable(ctx context.Context, item bo.Datasource, in bo.DropSqlTableInput, operatorUID string) (string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return "", err
	}
	statement := "DROP TABLE "
	if in.IfExists {
		statement += "IF EXISTS "
	}
	statement += sqlTableName(in.Schema, in.Table)
	statements := []string{statement}
	if err := r.execSqlDDL(ctx, item, in.Database, datasourcesqlaudit.ActionDROP_TABLE, sqlSchema(in.Schema), in.Table, statements, operatorUID); err != nil {
		return "", err
	}
	return statement, nil
}

// ListSqlAudits 按时间倒序返回某个数据源的 DDL 审计，并补齐操作人名称。
func (r *datasourceRepo) ListSqlAudits(ctx context.Context, namespaceUID, datasourceUID string, page, pageSize int) ([]bo.SqlAudit, int, error) {
	query := r.client.DatasourceSQLAudit.Query().Where(
		datasourcesqlaudit.NamespaceUIDEQ(namespaceUID),
		datasourcesqlaudit.DatasourceUIDEQ(datasourceUID),
	)
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(datasourcesqlaudit.FieldCreatedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	operatorUIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.OperatorUID != "" {
			operatorUIDs = append(operatorUIDs, row.OperatorUID)
		}
	}
	names := make(map[string]string, len(operatorUIDs))
	if len(operatorUIDs) > 0 {
		users, err := r.client.User.Query().Where(user.UIDIn(operatorUIDs...)).All(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		for _, account := range users {
			names[account.UID] = account.Name
		}
	}
	items := make([]bo.SqlAudit, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SqlAudit{
			UID: row.UID, Action: string(row.Action), Database: row.Database,
			Schema: row.Schema, Table: row.Table, Statement: row.Statement,
			Status: string(row.Status), ErrorMessage: row.ErrorMessage,
			OperatorUID: row.OperatorUID, OperatorName: names[row.OperatorUID],
			CreatedAt: row.CreatedAt,
		})
	}
	return items, total, nil
}

// execSqlDDL 在事务里执行结构变更，无论成功失败都写一条审计。
func (r *datasourceRepo) execSqlDDL(ctx context.Context, item bo.Datasource, database string, action datasourcesqlaudit.Action, schema, table string, statements []string, operatorUID string) error {
	err := r.runSqlDDL(ctx, item, database, statements)
	if auditErr := r.recordSqlAudit(ctx, item, action, database, schema, table, statements, operatorUID, err); auditErr != nil {
		r.log.Error(fmt.Sprintf("datasource %s sql audit: %v", item.UID, auditErr))
	}
	return err
}

func (r *datasourceRepo) runSqlDDL(ctx context.Context, item bo.Datasource, database string, statements []string) error {
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return err
	}
	defer db.Close()
	ddlCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	tx, err := db.BeginTx(ddlCtx, nil)
	if err != nil {
		return sqlClientError(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, statement := range statements {
		if _, err := tx.ExecContext(ddlCtx, statement); err != nil {
			return sqlClientError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return sqlClientError(err)
	}
	return nil
}

// recordSqlAudit 用独立上下文写审计，请求被取消时也能留下记录。
func (r *datasourceRepo) recordSqlAudit(ctx context.Context, item bo.Datasource, action datasourcesqlaudit.Action, database, schema, table string, statements []string, operatorUID string, runErr error) error {
	status := datasourcesqlaudit.StatusSUCCESS
	message := ""
	if runErr != nil {
		status = datasourcesqlaudit.StatusFAILED
		message = sqlAuditMessage(runErr)
	}
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), sqlAuditTimeout)
	defer cancel()
	return r.client.DatasourceSQLAudit.Create().
		SetUID(id.New()).
		SetNamespaceUID(item.NamespaceUID).
		SetDatasourceUID(item.UID).
		SetOperatorUID(operatorUID).
		SetAction(action).
		SetStatus(status).
		SetDatabase(database).
		SetSchema(schema).
		SetTable(table).
		SetStatement(strings.Join(statements, ";\n")).
		SetErrorMessage(message).
		Exec(auditCtx)
}

// sqlAuditMessage 取失败原因的第一行，按长度截断。
func sqlAuditMessage(err error) string {
	reason := strings.TrimSpace(err.Error())
	if index := strings.IndexByte(reason, '\n'); index >= 0 {
		reason = reason[:index]
	}
	runes := []rune(reason)
	if len(runes) > sqlMaxAuditError {
		reason = string(runes[:sqlMaxAuditError]) + "…"
	}
	return reason
}

// querySqlStrings 执行一条只返回一列文本的查询。
func (r *datasourceRepo) querySqlStrings(ctx context.Context, item bo.Datasource, database, statement string) ([]string, error) {
	db, err := r.sqlDataSourceFor(ctx, item, database)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	rows, err := db.QueryContext(queryCtx, statement)
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

// sqlSchema 把空的模式名归一成 public。
func sqlSchema(schema string) string {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return "public"
	}
	return schema
}

// sqlTableName 返回带模式的表名，标识符都加了双引号。
func sqlTableName(schema, table string) string {
	return quoteSqlIdentifier(sqlSchema(schema)) + "." + quoteSqlIdentifier(table)
}

// quoteSqlIdentifier 用双引号包裹标识符，内部双引号翻倍。
// 调用方必须先用 requireSqlName 校验，保证不会引入其它字符。
func quoteSqlIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func requireSqlName(name string) error {
	if !sqlNamePattern.MatchString(strings.TrimSpace(name)) {
		return merr.Invalid("invalid sql identifier")
	}
	return nil
}

// requireReadOnlySql 只允许读取类语句开头。真正的写保护由只读事务兜底。
func requireReadOnlySql(statement string) error {
	keyword := firstSqlKeyword(statement)
	if _, ok := sqlReadOnlyKeywords[keyword]; !ok {
		return merr.Invalid("only read-only statements are allowed")
	}
	return nil
}

// firstSqlKeyword 跳过空白和注释，返回第一个关键字的大写形式。
func firstSqlKeyword(statement string) string {
	rest := strings.TrimSpace(statement)
	for {
		switch {
		case strings.HasPrefix(rest, "--"):
			if index := strings.IndexByte(rest, '\n'); index >= 0 {
				rest = strings.TrimSpace(rest[index+1:])
				continue
			}
			return ""
		case strings.HasPrefix(rest, "/*"):
			if index := strings.Index(rest, "*/"); index >= 0 {
				rest = strings.TrimSpace(rest[index+2:])
				continue
			}
			return ""
		}
		break
	}
	end := strings.IndexFunc(rest, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z')
	})
	if end >= 0 {
		rest = rest[:end]
	}
	return strings.ToUpper(rest)
}

// sqlCellText 把一行单元格转成展示文本。NULL 为空串，时间用 RFC 3339。
func sqlCellText(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(typed)
	case string:
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case time.Time:
		return typed.Format(time.RFC3339Nano)
	case float64:
		return fmt.Sprintf("%v", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case int32:
		return fmt.Sprintf("%d", typed)
	case int:
		return fmt.Sprintf("%d", typed)
	default:
		return fmt.Sprint(typed)
	}
}

// validateSqlType 校验字段类型写法，返回去掉多余空白后的类型。
func validateSqlType(raw string) (string, error) {
	value := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if !sqlTypePattern.MatchString(value) {
		return "", merr.Invalid("invalid sql column type")
	}
	return strings.ToLower(value), nil
}

// validateSqlDefault 校验默认值写法。留空表示没有默认值。
func validateSqlDefault(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	for _, pattern := range sqlDefaultPatterns {
		if pattern.MatchString(value) {
			return value, nil
		}
	}
	return "", merr.Invalid("invalid sql default value")
}

// buildCreateSqlTable 生成建表语句。主键字段强制非空，字段名不允许重复。
func buildCreateSqlTable(in bo.CreateSqlTableInput) ([]string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return nil, err
	}
	if len(in.Columns) == 0 {
		return nil, merr.Invalid("columns are required")
	}
	definitions, err := buildSqlColumnDefinitions(in.Columns)
	if err != nil {
		return nil, err
	}
	primaryKeys := make([]string, 0, 1)
	statement := "CREATE TABLE "
	if in.IfNotExists {
		statement += "IF NOT EXISTS "
	}
	statement += sqlTableName(in.Schema, in.Table) + " (\n  " + strings.Join(definitions, ",\n  ")
	for _, column := range in.Columns {
		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, quoteSqlIdentifier(column.Name))
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
		statements = append(statements, "COMMENT ON TABLE "+sqlTableName(in.Schema, in.Table)+" IS "+quoteSqlLiteral(comment))
	}
	for _, column := range in.Columns {
		columnComment, err := validateSqlComment(column.Comment)
		if err != nil {
			return nil, err
		}
		if columnComment == "" {
			continue
		}
		statements = append(statements, "COMMENT ON COLUMN "+sqlTableName(in.Schema, in.Table)+"."+quoteSqlIdentifier(column.Name)+" IS "+quoteSqlLiteral(columnComment))
	}
	return statements, nil
}

// buildSqlColumnDefinitions 生成字段定义，每个字段名只允许出现一次。
func buildSqlColumnDefinitions(columns []bo.SqlColumnDefinition) ([]string, error) {
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
		columnType, err := validateSqlType(column.Type)
		if err != nil {
			return nil, err
		}
		defaultValue, err := validateSqlDefault(column.DefaultValue)
		if err != nil {
			return nil, err
		}
		definition := quoteSqlIdentifier(name) + " " + columnType
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

// buildAlterSqlTable 生成改表语句。至少要有一次修改，字段名不允许重复。
func buildAlterSqlTable(in bo.AlterSqlTableInput) ([]string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return nil, err
	}
	target := sqlTableName(in.Schema, in.Table)
	statements := make([]string, 0, 8)
	if in.RenameTo != "" {
		if err := requireSqlName(in.RenameTo); err != nil {
			return nil, err
		}
		if in.RenameTo == in.Table {
			return nil, merr.Invalid("invalid sql rename")
		}
		statements = append(statements, "ALTER TABLE "+target+" RENAME TO "+quoteSqlIdentifier(in.RenameTo))
	}
	if len(in.AddColumns) > 0 {
		definitions, err := buildSqlColumnDefinitions(in.AddColumns)
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
		statements = append(statements, "ALTER TABLE "+target+" DROP COLUMN "+quoteSqlIdentifier(name))
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
		quoted := quoteSqlIdentifier(name)
		changes := 0
		if column.Type != "" {
			columnType, err := validateSqlType(column.Type)
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
			defaultValue, err := validateSqlDefault(column.DefaultValue)
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
	if len(statements) == 0 {
		return nil, merr.Invalid("nothing to alter")
	}
	return statements, nil
}

// validateSqlComment 校验注释长度，注释以字面量形式写入。
func validateSqlComment(comment string) (string, error) {
	comment = strings.TrimSpace(comment)
	if utf8.RuneCountInString(comment) > 200 {
		return "", merr.Invalid("sql comment is too long")
	}
	return comment, nil
}

// quoteSqlLiteral 把文本转成 SQL 字符串字面量，单引号翻倍。
func quoteSqlLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
