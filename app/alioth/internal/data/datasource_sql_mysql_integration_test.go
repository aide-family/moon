package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasourcesqlaudit"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// TestMysqlLabBrowsesAndManagesTables 用真实 MySQL 覆盖库表浏览、只读查询、
// 表数据分页、建表改表删表和审计。需要 ALIOTH_TEST_DSN 指向的 PostgreSQL
// 测试库（存租户与审计）和 ALIOTH_TEST_MYSQL_DSN 指向的 MySQL 测试库；
// 任一未配置时跳过，绝不在开发库上执行 DDL。
func TestMysqlLabBrowsesAndManagesTables(t *testing.T) {
	source := testDSN(t)
	mysqlSource := strings.TrimSpace(os.Getenv("ALIOTH_TEST_MYSQL_DSN"))
	if mysqlSource == "" {
		t.Skip("set ALIOTH_TEST_MYSQL_DSN to a MySQL database other than the dev database")
	}
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

	namespace := "ns-mysql-lab-" + time.Now().Format("150405.000000")
	item := bo.Datasource{
		UID: "ds-mysql-" + time.Now().Format("150405.000000"), NamespaceUID: namespace,
		Name: "mysql", Type: "SQL", Driver: "SQL_MYSQL", Status: "ACTIVE",
		Endpoint: mysqlTestEndpoint(t, mysqlSource), Username: mysqlTestUser(t, mysqlSource),
		Password: mysqlTestPassword(t, mysqlSource),
	}
	table := fmt.Sprintf("t_mysql_lab_%d", time.Now().UnixNano()%1_000_000)
	t.Cleanup(func() {
		_, _ = repo.DropSqlTable(ctx, item, bo.DropSqlTableInput{
			SqlTableRef: bo.SqlTableRef{Table: table}, IfExists: true,
		}, "")
		_, _ = store.client.DatasourceSQLAudit.Delete().Where(datasourcesqlaudit.NamespaceUIDEQ(namespace)).Exec(ctx)
	})

	// 建表：反引号标识符、主键强制非空、注释内联。
	statement, err := repo.CreateSqlTable(ctx, item, bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Table: table},
		Columns: []bo.SqlColumnDefinition{
			{Name: "id", Type: "bigint", PrimaryKey: true},
			{Name: "name", Type: "VARCHAR(64)", Comment: "名称"},
			{Name: "created_at", Type: "datetime", DefaultValue: "CURRENT_TIMESTAMP"},
		},
		Comment: "MySQL 实验室测试表",
	}, "operator-1")
	if err != nil {
		t.Fatalf("CreateSqlTable: %v", err)
	}
	if !strings.Contains(statement, "`"+table+"`") || !strings.Contains(statement, "COMMENT '名称'") {
		t.Fatalf("create statement = %s", statement)
	}

	// 库表浏览能看到新表、字段和主键索引。
	tables, err := repo.SqlTables(ctx, item, "", "")
	if err != nil {
		t.Fatalf("SqlTables: %v", err)
	}
	if !containsSqlTable(tables, table) {
		t.Fatalf("created table %s missing", table)
	}
	columns, err := repo.SqlTableColumns(ctx, item, "", "", table)
	if err != nil {
		t.Fatalf("SqlTableColumns: %v", err)
	}
	if len(columns) != 3 || columns[0].Name != "id" || !columns[0].PrimaryKey || columns[0].Nullable {
		t.Fatalf("columns = %+v", columns)
	}
	if columns[1].DataType != "varchar(64)" || columns[1].UDTName != "varchar" {
		t.Fatalf("name column = %+v", columns[1])
	}
	if columns[2].DefaultValue != "CURRENT_TIMESTAMP" {
		t.Fatalf("created_at default = %q", columns[2].DefaultValue)
	}
	indexes, err := repo.SqlTableIndexes(ctx, item, "", "", table)
	if err != nil {
		t.Fatalf("SqlTableIndexes: %v", err)
	}
	if len(indexes) == 0 || !indexes[0].Primary || indexes[0].Definition != "PRIMARY KEY (`id`)" {
		t.Fatalf("indexes = %+v", indexes)
	}

	// 直接连 MySQL 写入两行，再验证只读查询和分页取数。
	writeDB, err := sql.Open("mysql", mysqlTestDSN(t, item))
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	defer writeDB.Close()
	if _, err := writeDB.ExecContext(ctx, fmt.Sprintf(
		"INSERT INTO `%s` (id, name) VALUES (1, 'alpha'), (2, 'beta')", table)); err != nil {
		t.Fatalf("insert rows: %v", err)
	}
	result, err := repo.QuerySql(ctx, item, "", fmt.Sprintf("SELECT id, name FROM `%s` ORDER BY id", table), 10)
	if err != nil {
		t.Fatalf("QuerySql: %v", err)
	}
	if len(result.Rows) != 2 || result.Rows[0][1] != "alpha" || result.Rows[1][1] != "beta" {
		t.Fatalf("query rows = %v", result.Rows)
	}
	// 写语句在只读入口被拒绝。
	if _, err := repo.QuerySql(ctx, item, "", fmt.Sprintf("DELETE FROM `%s`", table), 10); !merr.IsInvalid(err) {
		t.Fatalf("delete query err = %v", err)
	}

	// 分页取数：每页一行，第一页还有更多。
	firstPage, err := repo.SqlTableData(ctx, item, "", "", table, 1, 1)
	if err != nil {
		t.Fatalf("SqlTableData page 1: %v", err)
	}
	if len(firstPage.Rows) != 1 || !firstPage.HasMore || firstPage.Rows[0][1] != "alpha" {
		t.Fatalf("page 1 = %+v", firstPage)
	}
	secondPage, err := repo.SqlTableData(ctx, item, "", "", table, 2, 1)
	if err != nil {
		t.Fatalf("SqlTableData page 2: %v", err)
	}
	if len(secondPage.Rows) != 1 || secondPage.HasMore || secondPage.Rows[0][1] != "beta" {
		t.Fatalf("page 2 = %+v", secondPage)
	}

	// 改表：加字段、改可空、清默认值。MODIFY COLUMN 要保留没改的属性。
	nullable := true
	statements, err := repo.AlterSqlTable(ctx, item, bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Table: table},
		AddColumns:   []bo.SqlColumnDefinition{{Name: "remark", Type: "text", Nullable: true}},
		AlterColumns: []bo.SqlAlterColumn{{Name: "name", Nullable: &nullable}, {Name: "created_at", DropDefault: true}},
	}, "operator-1")
	if err != nil {
		t.Fatalf("AlterSqlTable: %v", err)
	}
	if len(statements) != 3 {
		t.Fatalf("alter statements = %v", statements)
	}
	columns, err = repo.SqlTableColumns(ctx, item, "", "", table)
	if err != nil {
		t.Fatalf("SqlTableColumns after alter: %v", err)
	}
	if len(columns) != 4 || !columns[1].Nullable || columns[2].DefaultValue != "" {
		t.Fatalf("columns after alter = %+v", columns)
	}

	// 重复建表失败也要留下一条失败审计；MySQL 没有事务 DDL，审计记录尝试过的语句。
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
	tables, err = repo.SqlTables(ctx, item, "", "")
	if err != nil {
		t.Fatalf("SqlTables after drop: %v", err)
	}
	if containsSqlTable(tables, table) {
		t.Fatalf("dropped table %s still listed", table)
	}
}

// mysqlTestEndpoint 把测试库 DSN 换成不带凭据的 mysql:// 连接串，
// 凭据走数据源的用户名密码字段。
func mysqlTestEndpoint(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse mysql test dsn: %v", err)
	}
	parsed.User = nil
	parsed.RawQuery = ""
	return parsed.String()
}

func mysqlTestUser(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.User == nil {
		return ""
	}
	return parsed.User.Username()
}

func mysqlTestPassword(t *testing.T, dsn string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.User == nil {
		return ""
	}
	password, _ := parsed.User.Password()
	return password
}

// mysqlTestDSN 用数据源配置生成数据库驱动连接串，供测试直接写入数据。
func mysqlTestDSN(t *testing.T, item bo.Datasource) string {
	t.Helper()
	dsn, err := mysqlDialect{}.dsn(item, "")
	if err != nil {
		t.Fatalf("mysql dsn: %v", err)
	}
	return dsn
}
