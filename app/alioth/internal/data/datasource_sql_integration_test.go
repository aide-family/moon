package data

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasourcesqlaudit"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// TestSqlLabBrowsesAndManagesTables 用真实 PostgreSQL 覆盖库表浏览、只读查询、
// 建表改表删表和审计。数据库地址取 ALIOTH_TEST_DSN 指向的测试库，
// 没有配置时跳过，绝不在开发库上执行 DDL。
func TestSqlLabBrowsesAndManagesTables(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	repo := NewDatasourceRepo(store, slog.Default())

	namespace := "ns-sql-lab-" + time.Now().Format("150405.000000")
	item := bo.Datasource{
		UID: "ds-sql-" + time.Now().Format("150405.000000"), NamespaceUID: namespace,
		Name: "pg", Type: "SQL", Driver: "SQL_POSTGRESQL", Status: "ACTIVE",
		Endpoint: sqlTestEndpoint(t, source), Username: sqlTestUser(t, source),
		Password: sqlTestPassword(t, source),
	}
	table := fmt.Sprintf("t_sql_lab_%d", time.Now().UnixNano()%1_000_000)
	t.Cleanup(func() {
		_, _ = repo.DropSqlTable(ctx, item, bo.DropSqlTableInput{
			SqlTableRef: bo.SqlTableRef{Table: table}, IfExists: true,
		}, "")
		_, _ = store.client.DatasourceSQLAudit.Delete().Where(datasourcesqlaudit.NamespaceUIDEQ(namespace)).Exec(ctx)
	})

	// 建表：主键强制非空，字段注释单独一条语句。
	statement, err := repo.CreateSqlTable(ctx, item, bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Table: table},
		Columns: []bo.SqlColumnDefinition{
			{Name: "id", Type: "bigint", PrimaryKey: true},
			{Name: "name", Type: "VARCHAR(64)", Comment: "名称"},
			{Name: "created_at", Type: "timestamptz", DefaultValue: "now()"},
		},
		Comment: "SQL 实验室测试表",
	}, "operator-1")
	if err != nil {
		t.Fatalf("CreateSqlTable: %v", err)
	}
	if statement == "" {
		t.Fatal("empty create statement")
	}

	// 库表浏览能看到新表、字段和主键索引。
	tables, err := repo.SqlTables(ctx, item, "", "public")
	if err != nil {
		t.Fatalf("SqlTables: %v", err)
	}
	if !containsSqlTable(tables, table) {
		t.Fatalf("created table %s missing from %v", table, tables)
	}
	columns, err := repo.SqlTableColumns(ctx, item, "", "public", table)
	if err != nil {
		t.Fatalf("SqlTableColumns: %v", err)
	}
	if len(columns) != 3 || columns[0].Name != "id" || !columns[0].PrimaryKey || columns[0].Nullable {
		t.Fatalf("columns = %+v", columns)
	}
	if columns[1].DataType != "character varying" || columns[1].Length != 64 {
		t.Fatalf("name column = %+v", columns[1])
	}
	if columns[2].DefaultValue == "" {
		t.Fatalf("created_at default missing: %+v", columns[2])
	}
	indexes, err := repo.SqlTableIndexes(ctx, item, "", "public", table)
	if err != nil {
		t.Fatalf("SqlTableIndexes: %v", err)
	}
	if len(indexes) == 0 || !indexes[0].Primary {
		t.Fatalf("indexes = %+v", indexes)
	}

	// 只读查询返回列名和字符串化的行。
	if _, err := store.db.ExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (id, name) VALUES (1, 'alpha'), (2, 'beta')`, quoteSqlIdentifier(table))); err != nil {
		t.Fatalf("insert rows: %v", err)
	}
	result, err := repo.QuerySql(ctx, item, "", fmt.Sprintf(`SELECT id, name, created_at FROM %s ORDER BY id`, quoteSqlIdentifier(table)), 10)
	if err != nil {
		t.Fatalf("QuerySql: %v", err)
	}
	if len(result.Columns) != 3 || result.Columns[0] != "id" || result.Columns[1] != "name" {
		t.Fatalf("query columns = %v", result.Columns)
	}
	if len(result.Rows) != 2 || result.Rows[0][0] != "1" || result.Rows[0][1] != "alpha" || result.Rows[1][1] != "beta" {
		t.Fatalf("query rows = %v", result.Rows)
	}
	if result.Rows[0][2] == "" || result.Truncated {
		t.Fatalf("timestamp cell = %q truncated = %v", result.Rows[0][2], result.Truncated)
	}
	// 表数据分页：默认按主键排序，每页取一行，最后一页 hasMore 为 false。
	firstPage, err := repo.SqlTableData(ctx, item, "", "public", table, 1, 1)
	if err != nil {
		t.Fatalf("SqlTableData page 1: %v", err)
	}
	if len(firstPage.Columns) != 3 || len(firstPage.Rows) != 1 || !firstPage.HasMore {
		t.Fatalf("page 1 = %+v", firstPage)
	}
	if firstPage.Rows[0][0] != "1" || firstPage.Rows[0][1] != "alpha" {
		t.Fatalf("page 1 rows = %v", firstPage.Rows)
	}
	secondPage, err := repo.SqlTableData(ctx, item, "", "public", table, 2, 1)
	if err != nil {
		t.Fatalf("SqlTableData page 2: %v", err)
	}
	if len(secondPage.Rows) != 1 || secondPage.HasMore || secondPage.Rows[0][1] != "beta" {
		t.Fatalf("page 2 = %+v", secondPage)
	}
	emptyPage, err := repo.SqlTableData(ctx, item, "", "public", table, 3, 1)
	if err != nil {
		t.Fatalf("SqlTableData page 3: %v", err)
	}
	if len(emptyPage.Rows) != 0 || emptyPage.HasMore {
		t.Fatalf("page 3 = %+v", emptyPage)
	}

	// 行数上限截断时标记 truncated。
	truncated, err := repo.QuerySql(ctx, item, "", fmt.Sprintf(`SELECT id FROM %s ORDER BY id`, quoteSqlIdentifier(table)), 1)
	if err != nil {
		t.Fatalf("QuerySql truncated: %v", err)
	}
	if len(truncated.Rows) != 1 || !truncated.Truncated {
		t.Fatalf("truncated = %+v", truncated)
	}

	// 写语句在只读入口被拒绝。
	if _, err := repo.QuerySql(ctx, item, "", fmt.Sprintf(`DELETE FROM %s`, quoteSqlIdentifier(table)), 10); !merr.IsInvalid(err) {
		t.Fatalf("delete query err = %v", err)
	}

	// 改表：加字段、改类型、清默认值。
	nullable := true
	statements, err := repo.AlterSqlTable(ctx, item, bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Table: table},
		AddColumns:   []bo.SqlColumnDefinition{{Name: "remark", Type: "text", Nullable: true}},
		AlterColumns: []bo.SqlAlterColumn{{Name: "name", Type: "varchar(128)", Nullable: &nullable}},
	}, "operator-1")
	if err != nil {
		t.Fatalf("AlterSqlTable: %v", err)
	}
	if len(statements) != 3 {
		t.Fatalf("alter statements = %v", statements)
	}
	columns, err = repo.SqlTableColumns(ctx, item, "", "public", table)
	if err != nil {
		t.Fatalf("SqlTableColumns after alter: %v", err)
	}
	if len(columns) != 4 || columns[1].Length != 128 || !columns[1].Nullable {
		t.Fatalf("columns after alter = %+v", columns)
	}

	// 重复建表失败也要留下一条失败审计。
	if _, err := repo.CreateSqlTable(ctx, item, bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Table: table},
		Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "bigint"}},
	}, "operator-2"); err == nil {
		t.Fatal("duplicate create should fail")
	}
	audits, total, err := repo.ListSqlAudits(ctx, namespace, item.UID, 1, 20)
	if err != nil {
		t.Fatalf("ListSqlAudits: %v", err)
	}
	if total != 3 || len(audits) != 3 {
		t.Fatalf("audits total = %d len = %d", total, len(audits))
	}
	failed := 0
	for _, audit := range audits {
		if audit.Status == "FAILED" {
			failed++
			if audit.ErrorMessage == "" || audit.OperatorUID != "operator-2" {
				t.Fatalf("failed audit = %+v", audit)
			}
		}
	}
	if failed != 1 {
		t.Fatalf("failed audits = %d", failed)
	}

	// 删表后浏览结果里不再有它。
	if _, err := repo.DropSqlTable(ctx, item, bo.DropSqlTableInput{SqlTableRef: bo.SqlTableRef{Table: table}}, "operator-1"); err != nil {
		t.Fatalf("DropSqlTable: %v", err)
	}
	tables, err = repo.SqlTables(ctx, item, "", "public")
	if err != nil {
		t.Fatalf("SqlTables after drop: %v", err)
	}
	if containsSqlTable(tables, table) {
		t.Fatalf("dropped table %s still listed", table)
	}
}

func containsSqlTable(tables []bo.SqlTable, name string) bool {
	for _, table := range tables {
		if table.Name == name {
			return true
		}
	}
	return false
}

// sqlTestEndpoint 把测试库 DSN 换成不带凭据的连接串，凭据走数据源的用户名密码字段。
func sqlTestEndpoint(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse test dsn: %v", err)
	}
	parsed.User = nil
	parsed.Scheme = "postgres"
	parsed.RawQuery = "sslmode=disable"
	return parsed.String()
}

func sqlTestUser(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse test dsn: %v", err)
	}
	if parsed.User == nil {
		return ""
	}
	return parsed.User.Username()
}

func sqlTestPassword(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse test dsn: %v", err)
	}
	if parsed.User == nil {
		return ""
	}
	password, _ := parsed.User.Password()
	return password
}
