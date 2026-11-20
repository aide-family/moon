package data

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
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

// sqlNamePattern 限定库名、模式名、表名和字段名。只接受不加引号也安全的
// 标识符，避免把用户输入拼进 DDL 时出现注入。PostgreSQL 和 MySQL 的标识符
// 规则在这条子集上一致。
var sqlNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

// sqlReadOnlyKeywords 是只读查询允许的首个关键字。
var sqlReadOnlyKeywords = map[string]struct{}{
	"SELECT":  {},
	"WITH":    {},
	"SHOW":    {},
	"EXPLAIN": {},
	"VALUES":  {},
	"TABLE":   {},
	"DESC":    {},
	"DESCRIBE": {},
}

// sqlDataSourceFor 打开用户数据库的连接。每次请求只开一个连接，
// 用完即关，不在进程里缓存用户数据库的会话。
func (r *datasourceRepo) sqlDataSourceFor(ctx context.Context, item bo.Datasource, database string, dialect sqlDialect) (*sql.DB, error) {
	dsn, err := dialect.dsn(item, database)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(dialect.driverName(), dsn)
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
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, "", dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	return dialect.databases(queryCtx, db)
}

// SqlSchemas 列出库里的用户模式。MySQL 里模式和库是同一层。
func (r *datasourceRepo) SqlSchemas(ctx context.Context, item bo.Datasource, database string) ([]string, error) {
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database, dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	return dialect.schemas(queryCtx, db, database)
}

// SqlTables 列出模式里的表和视图。
func (r *datasourceRepo) SqlTables(ctx context.Context, item bo.Datasource, database, schema string) ([]bo.SqlTable, error) {
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database, dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	return dialect.tables(queryCtx, db, dialect.schemaName(schema, database))
}

// SqlTableColumns 列出表的字段、默认值、主键和唯一约束。
func (r *datasourceRepo) SqlTableColumns(ctx context.Context, item bo.Datasource, database, schema, table string) ([]bo.SqlColumn, error) {
	if err := requireSqlName(table); err != nil {
		return nil, err
	}
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database, dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	return dialect.columns(queryCtx, db, dialect.schemaName(schema, database), table)
}

// SqlTableIndexes 列出表的索引。
func (r *datasourceRepo) SqlTableIndexes(ctx context.Context, item bo.Datasource, database, schema, table string) ([]bo.SqlIndex, error) {
	if err := requireSqlName(table); err != nil {
		return nil, err
	}
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database, dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	return dialect.indexes(queryCtx, db, dialect.schemaName(schema, database), table)
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
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database, dialect)
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

// SqlTableData 分页读取表数据。排序用主键；PostgreSQL 没有主键时按物理行号
// ctid 排序，MySQL 没有主键时使用存储顺序。
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
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, database, dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	queryCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	effectiveSchema := dialect.schemaName(schema, database)
	primaryKeys, err := dialect.primaryKeys(queryCtx, db, effectiveSchema, table)
	if err != nil {
		return nil, err
	}
	// 多取一行用来判断是否还有下一页。
	statement, err := dialect.tableDataStatement(effectiveSchema, table, primaryKeys, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
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

// CreateSqlTable 建表并写审计。
func (r *datasourceRepo) CreateSqlTable(ctx context.Context, item bo.Datasource, in bo.CreateSqlTableInput, operatorUID string) (string, error) {
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return "", err
	}
	db, err := r.sqlDataSourceFor(ctx, item, in.Database, dialect)
	if err != nil {
		return "", err
	}
	defer db.Close()
	schema := dialect.schemaName(in.Schema, in.Database)
	statements, err := dialect.createStatements(ctx, db, in, schema)
	if err != nil {
		return "", err
	}
	if err := r.execSqlDDL(ctx, db, dialect, item, in.Database, datasourcesqlaudit.ActionCREATE_TABLE, schema, in.Table, statements, operatorUID); err != nil {
		return "", err
	}
	return strings.Join(statements, ";\n"), nil
}

// AlterSqlTable 改表并写审计。rename、加字段、删字段和改字段可以一次提交。
func (r *datasourceRepo) AlterSqlTable(ctx context.Context, item bo.Datasource, in bo.AlterSqlTableInput, operatorUID string) ([]string, error) {
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return nil, err
	}
	db, err := r.sqlDataSourceFor(ctx, item, in.Database, dialect)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	schema := dialect.schemaName(in.Schema, in.Database)
	statements, err := dialect.alterStatements(ctx, db, in, schema)
	if err != nil {
		return nil, err
	}
	if err := r.execSqlDDL(ctx, db, dialect, item, in.Database, datasourcesqlaudit.ActionALTER_TABLE, schema, in.Table, statements, operatorUID); err != nil {
		return nil, err
	}
	return statements, nil
}

// DropSqlTable 删表并写审计。
func (r *datasourceRepo) DropSqlTable(ctx context.Context, item bo.Datasource, in bo.DropSqlTableInput, operatorUID string) (string, error) {
	if err := requireSqlName(in.Table); err != nil {
		return "", err
	}
	dialect, err := sqlDialectFor(item.Driver)
	if err != nil {
		return "", err
	}
	db, err := r.sqlDataSourceFor(ctx, item, in.Database, dialect)
	if err != nil {
		return "", err
	}
	defer db.Close()
	schema := dialect.schemaName(in.Schema, in.Database)
	statement, err := dialect.dropStatement(in, schema)
	if err != nil {
		return "", err
	}
	statements := []string{statement}
	if err := r.execSqlDDL(ctx, db, dialect, item, in.Database, datasourcesqlaudit.ActionDROP_TABLE, schema, in.Table, statements, operatorUID); err != nil {
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
func (r *datasourceRepo) execSqlDDL(ctx context.Context, db *sql.DB, dialect sqlDialect, item bo.Datasource, database string, action datasourcesqlaudit.Action, schema, table string, statements []string, operatorUID string) error {
	err := r.runSqlDDL(ctx, db, dialect, statements)
	if auditErr := r.recordSqlAudit(ctx, item, action, database, schema, table, statements, operatorUID, err); auditErr != nil {
		r.log.Error(fmt.Sprintf("datasource %s sql audit: %v", item.UID, auditErr))
	}
	return err
}

func (r *datasourceRepo) runSqlDDL(ctx context.Context, db *sql.DB, dialect sqlDialect, statements []string) error {
	ddlCtx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()
	// MySQL 的 DDL 会隐式提交，包事务没有意义，按顺序执行并在第一条失败时停下。
	if !dialect.supportsDDLTransaction() {
		for _, statement := range statements {
			if _, err := db.ExecContext(ddlCtx, statement); err != nil {
				return sqlClientError(err)
			}
		}
		return nil
	}
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
		case strings.HasPrefix(rest, "#"):
			if index := strings.IndexByte(rest, '\n'); index >= 0 {
				rest = strings.TrimSpace(rest[index+1:])
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
	case float32:
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

// validateSqlComment 校验注释长度，注释以字面量形式写入。
func validateSqlComment(comment string) (string, error) {
	comment = strings.TrimSpace(comment)
	if utf8.RuneCountInString(comment) > 200 {
		return "", merr.Invalid("sql comment is too long")
	}
	return comment, nil
}

// quoteSqlLiteral 把文本转成 SQL 字符串字面量，单引号翻倍。两种数据库一致。
func quoteSqlLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
