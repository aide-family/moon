package data

import (
	"context"
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func TestPostgresCreateStatements(t *testing.T) {
	statements, err := postgresDialect{}.createStatements(context.Background(), nil, bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Schema: "public", Table: "devices"},
		Columns: []bo.SqlColumnDefinition{
			{Name: "id", Type: "bigint", PrimaryKey: true},
			{Name: "name", Type: " VARCHAR(64) ", Comment: "设备名"},
			{Name: "created_at", Type: "timestamptz", DefaultValue: "now()"},
		},
		Comment: "设备",
	}, "public")
	if err != nil {
		t.Fatalf("createStatements: %v", err)
	}
	if len(statements) != 3 {
		t.Fatalf("statements = %v", statements)
	}
	create := statements[0]
	for _, want := range []string{`"public"."devices"`, `"id" bigint NOT NULL`, `"name" varchar(64) NOT NULL`, `"created_at" timestamptz NOT NULL DEFAULT now()`, `PRIMARY KEY ("id")`} {
		if !strings.Contains(create, want) {
			t.Fatalf("create statement missing %q: %s", want, create)
		}
	}
	if !strings.Contains(statements[1], `COMMENT ON TABLE "public"."devices" IS '设备'`) {
		t.Fatalf("table comment = %s", statements[1])
	}
	if !strings.Contains(statements[2], `COMMENT ON COLUMN "public"."devices"."name" IS '设备名'`) {
		t.Fatalf("column comment = %s", statements[2])
	}
}

func TestPostgresCreateStatementsRejectUnsafeInput(t *testing.T) {
	cases := map[string]bo.CreateSqlTableInput{
		"duplicate column": {
			SqlTableRef: bo.SqlTableRef{Table: "devices"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "bigint"}, {Name: "id", Type: "text"}},
		},
		"bad identifier": {
			SqlTableRef: bo.SqlTableRef{Table: "devices; drop table users"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "bigint"}},
		},
		"bad type": {
			SqlTableRef: bo.SqlTableRef{Table: "devices"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "int); drop table users; --"}},
		},
		"bad default": {
			SqlTableRef: bo.SqlTableRef{Table: "devices"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "text", DefaultValue: "'x'; drop table users"}},
		},
		"no column": {
			SqlTableRef: bo.SqlTableRef{Table: "devices"},
		},
	}
	for name, in := range cases {
		if _, err := (postgresDialect{}).createStatements(context.Background(), nil, in, "public"); !merr.IsInvalid(err) {
			t.Fatalf("%s: err = %v", name, err)
		}
	}
}

func TestPostgresAlterStatements(t *testing.T) {
	nullable := true
	statements, err := postgresDialect{}.alterStatements(context.Background(), nil, bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Schema: "ops", Table: "devices"},
		RenameTo:     "device",
		AddColumns:   []bo.SqlColumnDefinition{{Name: "owner", Type: "text", Nullable: true}},
		DropColumns:  []string{"legacy"},
		AlterColumns: []bo.SqlAlterColumn{{Name: "name", Type: "varchar(128)", Nullable: &nullable}},
	}, "ops")
	if err != nil {
		t.Fatalf("alterStatements: %v", err)
	}
	joined := strings.Join(statements, "\n")
	for _, want := range []string{
		`ALTER TABLE "ops"."devices" RENAME TO "device"`,
		`ALTER TABLE "ops"."devices" ADD COLUMN "owner" text`,
		`ALTER TABLE "ops"."devices" DROP COLUMN "legacy"`,
		`ALTER TABLE "ops"."devices" ALTER COLUMN "name" TYPE varchar(128)`,
		`ALTER TABLE "ops"."devices" ALTER COLUMN "name" DROP NOT NULL`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("alter statements missing %q: %s", want, joined)
		}
	}
	if strings.Contains(joined, `"owner" text NOT NULL`) {
		t.Fatalf("nullable column should not be NOT NULL: %s", joined)
	}
	// 改名必须放在最后，否则后续语句会指向已经改名的表。
	if !strings.Contains(statements[len(statements)-1], "RENAME TO") {
		t.Fatalf("rename should be the last statement: %v", statements)
	}
}

func TestPostgresAlterStatementsRequireChange(t *testing.T) {
	if _, err := (postgresDialect{}).alterStatements(context.Background(), nil, bo.AlterSqlTableInput{SqlTableRef: bo.SqlTableRef{Table: "devices"}}, "public"); !merr.IsInvalid(err) {
		t.Fatalf("empty alter err = %v", err)
	}
	if _, err := (postgresDialect{}).alterStatements(context.Background(), nil, bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Table: "devices"},
		AlterColumns: []bo.SqlAlterColumn{{Name: "name"}},
	}, "public"); !merr.IsInvalid(err) {
		t.Fatalf("alter without change err = %v", err)
	}
}

func TestRequireReadOnlySql(t *testing.T) {
	valid := []string{
		"select 1",
		"  SELECT * FROM devices",
		"-- 统计设备\nSELECT count(*) FROM devices",
		"/* hint */ WITH recent AS (SELECT 1) SELECT * FROM recent",
		"EXPLAIN SELECT 1",
		"SHOW search_path",
		"TABLE devices",
		"VALUES (1), (2)",
		"DESC devices",
	}
	for _, statement := range valid {
		if err := requireReadOnlySql(statement); err != nil {
			t.Fatalf("requireReadOnlySql(%q) = %v", statement, err)
		}
	}
	invalid := []string{
		"",
		"delete from devices",
		"INSERT INTO devices VALUES (1)",
		"update devices set name = 'x'",
		"DROP TABLE devices",
		"create table x (id int)",
		"-- only a comment",
	}
	for _, statement := range invalid {
		if err := requireReadOnlySql(statement); !merr.IsInvalid(err) {
			t.Fatalf("requireReadOnlySql(%q) = %v", statement, err)
		}
	}
}

func TestPostgresDSN(t *testing.T) {
	dsn, err := postgresDialect{}.dsn(bo.Datasource{
		Endpoint: "postgres://db.moon.com:5432/app",
		Username: "moon",
		Password: "p@ss/w0rd",
	}, "analytics")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	for _, want := range []string{"postgres://", "db.moon.com:5432/analytics", "application_name=alioth"} {
		if !strings.Contains(dsn, want) {
			t.Fatalf("dsn missing %q: %s", want, dsn)
		}
	}
	if strings.Contains(dsn, "p@ss/w0rd") {
		t.Fatalf("dsn should escape the password: %s", dsn)
	}
	if _, err := (postgresDialect{}).dsn(bo.Datasource{Endpoint: "http://db.moon.com"}, ""); !merr.IsInvalid(err) {
		t.Fatalf("http endpoint err = %v", err)
	}
	if _, err := (postgresDialect{}).dsn(bo.Datasource{Endpoint: "postgres://db.moon.com/app"}, "bad/name"); !merr.IsInvalid(err) {
		t.Fatalf("bad database err = %v", err)
	}
}

func TestSqlCellText(t *testing.T) {
	cases := []struct {
		value any
		want  string
	}{
		{nil, ""},
		{int64(3), "3"},
		{[]byte("abc"), "abc"},
		{true, "true"},
		{false, "false"},
	}
	for _, item := range cases {
		if got := sqlCellText(item.value); got != item.want {
			t.Fatalf("sqlCellText(%v) = %q, want %q", item.value, got, item.want)
		}
	}
}

func TestPostgresTableDataStatement(t *testing.T) {
	statement, err := postgresDialect{}.tableDataStatement("public", "devices", []string{"id", "region"}, 101, 200)
	if err != nil {
		t.Fatalf("tableDataStatement: %v", err)
	}
	want := `SELECT * FROM "public"."devices" ORDER BY "id", "region" LIMIT 101 OFFSET 200`
	if statement != want {
		t.Fatalf("statement = %s, want %s", statement, want)
	}
	// 没有主键时按物理行号排序，保证翻页有确定的顺序。
	statement, err = postgresDialect{}.tableDataStatement("ops", "events", nil, 100, 0)
	if err != nil {
		t.Fatalf("tableDataStatement: %v", err)
	}
	want = `SELECT * FROM "ops"."events" ORDER BY ctid LIMIT 100 OFFSET 0`
	if statement != want {
		t.Fatalf("statement = %s, want %s", statement, want)
	}
}

func TestMysqlDSN(t *testing.T) {
	dsn, err := mysqlDialect{}.dsn(bo.Datasource{
		Endpoint: "mysql://db.moon.com:3306/app",
		Username: "moon",
		Password: "p@ss/w0rd",
	}, "analytics")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	for _, want := range []string{"moon:", "@tcp(db.moon.com:3306)/analytics", "charset=utf8mb4", "parseTime=true"} {
		if !strings.Contains(dsn, want) {
			t.Fatalf("dsn missing %q: %s", want, dsn)
		}
	}
	// 连接串里的 userinfo 在没有显式字段时兜底。
	dsn, err = mysqlDialect{}.dsn(bo.Datasource{Endpoint: "mysql://root:secret@127.0.0.1:3306/app"}, "")
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	if !strings.Contains(dsn, "root:secret@tcp(127.0.0.1:3306)/app") {
		t.Fatalf("dsn = %s", dsn)
	}
	if _, err := (mysqlDialect{}).dsn(bo.Datasource{Endpoint: "postgres://db.moon.com/app"}, ""); !merr.IsInvalid(err) {
		t.Fatalf("postgres endpoint err = %v", err)
	}
	if _, err := (mysqlDialect{}).dsn(bo.Datasource{Endpoint: "mysql://db.moon.com:3306"}, ""); !merr.IsInvalid(err) {
		t.Fatalf("missing database err = %v", err)
	}
	if _, err := (mysqlDialect{}).dsn(bo.Datasource{Endpoint: "mysql://db.moon.com/app"}, "bad/name"); !merr.IsInvalid(err) {
		t.Fatalf("bad database err = %v", err)
	}
}

func TestMysqlCreateStatements(t *testing.T) {
	statements, err := mysqlDialect{}.createStatements(context.Background(), nil, bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Schema: "app", Table: "devices"},
		Columns: []bo.SqlColumnDefinition{
			{Name: "id", Type: "bigint", PrimaryKey: true},
			{Name: "seq", Type: "int unsigned", Unique: true},
			{Name: "name", Type: "varchar(64)", Comment: "设备名"},
			{Name: "created_at", Type: "datetime", DefaultValue: "CURRENT_TIMESTAMP"},
		},
		Comment: "设备",
	}, "app")
	if err != nil {
		t.Fatalf("createStatements: %v", err)
	}
	if len(statements) != 1 {
		t.Fatalf("statements = %v", statements)
	}
	create := statements[0]
	for _, want := range []string{
		"CREATE TABLE `app`.`devices`",
		"`id` bigint NOT NULL",
		"`seq` int unsigned NOT NULL UNIQUE",
		"`name` varchar(64) NOT NULL COMMENT '设备名'",
		"`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP",
		"PRIMARY KEY (`id`)",
		"COMMENT='设备'",
	} {
		if !strings.Contains(create, want) {
			t.Fatalf("create statement missing %q: %s", want, create)
		}
	}
}

func TestMysqlCreateStatementsRejectUnsafeInput(t *testing.T) {
	cases := map[string]bo.CreateSqlTableInput{
		"bad type": {
			SqlTableRef: bo.SqlTableRef{Table: "devices"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "int; drop table users"}},
		},
		"bad default": {
			SqlTableRef: bo.SqlTableRef{Table: "devices"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "text", DefaultValue: "'x'; drop table users"}},
		},
		"bad identifier": {
			SqlTableRef: bo.SqlTableRef{Table: "devices`; drop table users"},
			Columns:     []bo.SqlColumnDefinition{{Name: "id", Type: "bigint"}},
		},
	}
	for name, in := range cases {
		if _, err := (mysqlDialect{}).createStatements(context.Background(), nil, in, "app"); !merr.IsInvalid(err) {
			t.Fatalf("%s: err = %v", name, err)
		}
	}
}

func TestMysqlAlterStatementsAddDropRename(t *testing.T) {
	statements, err := mysqlDialect{}.alterStatements(context.Background(), nil, bo.AlterSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Schema: "app", Table: "devices"},
		RenameTo:    "device",
		AddColumns:  []bo.SqlColumnDefinition{{Name: "owner", Type: "varchar(64)", Nullable: true}},
		DropColumns: []string{"legacy"},
	}, "app")
	if err != nil {
		t.Fatalf("alterStatements: %v", err)
	}
	joined := strings.Join(statements, "\n")
	for _, want := range []string{
		"ALTER TABLE `app`.`devices` ADD COLUMN `owner` varchar(64)",
		"ALTER TABLE `app`.`devices` DROP COLUMN `legacy`",
		"ALTER TABLE `app`.`devices` RENAME TO `device`",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("alter statements missing %q: %s", want, joined)
		}
	}
	if !strings.Contains(statements[len(statements)-1], "RENAME TO") {
		t.Fatalf("rename should be the last statement: %v", statements)
	}
}

func TestMysqlModifyDefinitionKeepsExistingParts(t *testing.T) {
	nullable := false
	definition, err := mysqlDialect{}.modifyDefinition(bo.SqlAlterColumn{Name: "name", Nullable: &nullable}, mysqlColumnState{
		columnType: "varchar(64)", charset: "utf8mb4", collation: "utf8mb4_general_ci",
		defaultValue: "x", comment: "设备名",
	})
	if err != nil {
		t.Fatalf("modifyDefinition: %v", err)
	}
	// 文字默认值要补引号；没改的类型、字符集、注释和可空性都要保留。
	for _, want := range []string{"`name` varchar(64)", "CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", "NOT NULL", "DEFAULT 'x'", "COMMENT '设备名'"} {
		if !strings.Contains(definition, want) {
			t.Fatalf("definition missing %q: %s", want, definition)
		}
	}
	// 表达式默认值原样保留；清除默认值时不写 DEFAULT；自增属性保留。
	definition, err = mysqlDialect{}.modifyDefinition(bo.SqlAlterColumn{Name: "id", DropDefault: true}, mysqlColumnState{
		columnType: "bigint", extra: "auto_increment", defaultValue: "1",
	})
	if err != nil {
		t.Fatalf("modifyDefinition: %v", err)
	}
	if strings.Contains(definition, "DEFAULT") || !strings.Contains(definition, "AUTO_INCREMENT") {
		t.Fatalf("definition = %s", definition)
	}
	definition, err = mysqlDialect{}.modifyDefinition(bo.SqlAlterColumn{Name: "created_at"}, mysqlColumnState{
		columnType: "timestamp", extra: "DEFAULT_GENERATED", defaultValue: "CURRENT_TIMESTAMP",
	})
	if err != nil {
		t.Fatalf("modifyDefinition: %v", err)
	}
	if !strings.Contains(definition, "DEFAULT CURRENT_TIMESTAMP") {
		t.Fatalf("definition = %s", definition)
	}
}

func TestMysqlTableDataStatement(t *testing.T) {
	statement, err := mysqlDialect{}.tableDataStatement("app", "devices", []string{"id"}, 101, 200)
	if err != nil {
		t.Fatalf("tableDataStatement: %v", err)
	}
	want := "SELECT * FROM `app`.`devices` ORDER BY `id` LIMIT 101 OFFSET 200"
	if statement != want {
		t.Fatalf("statement = %s, want %s", statement, want)
	}
	// 没有主键时使用存储顺序，不拼 ORDER BY。
	statement, err = mysqlDialect{}.tableDataStatement("", "devices", nil, 100, 0)
	if err != nil {
		t.Fatalf("tableDataStatement: %v", err)
	}
	want = "SELECT * FROM `devices` LIMIT 100 OFFSET 0"
	if statement != want {
		t.Fatalf("statement = %s, want %s", statement, want)
	}
}

func TestSqlDialectFor(t *testing.T) {
	for _, driver := range []string{driverSQLPostgresql, driverSQLMysql} {
		if !isSqlDriver(driver) {
			t.Fatalf("isSqlDriver(%q) = false", driver)
		}
		if _, err := sqlDialectFor(driver); err != nil {
			t.Fatalf("sqlDialectFor(%q) = %v", driver, err)
		}
	}
	if isSqlDriver("METRICS_PROMETHEUS") {
		t.Fatal("metrics driver treated as sql")
	}
	if _, err := sqlDialectFor("PROBE_HTTP"); !merr.IsInvalid(err) {
		t.Fatalf("probe driver err = %v", err)
	}
}
