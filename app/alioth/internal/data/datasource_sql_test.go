package data

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func TestBuildCreateSqlTable(t *testing.T) {
	statements, err := buildCreateSqlTable(bo.CreateSqlTableInput{
		SqlTableRef: bo.SqlTableRef{Schema: "public", Table: "devices"},
		Columns: []bo.SqlColumnDefinition{
			{Name: "id", Type: "bigint", PrimaryKey: true},
			{Name: "name", Type: " VARCHAR(64) ", Comment: "设备名"},
			{Name: "created_at", Type: "timestamptz", DefaultValue: "now()"},
		},
		Comment: "设备",
	})
	if err != nil {
		t.Fatalf("buildCreateSqlTable: %v", err)
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

func TestBuildCreateSqlTableRejectsUnsafeInput(t *testing.T) {
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
		if _, err := buildCreateSqlTable(in); !merr.IsInvalid(err) {
			t.Fatalf("%s: err = %v", name, err)
		}
	}
}

func TestBuildAlterSqlTable(t *testing.T) {
	nullable := true
	statements, err := buildAlterSqlTable(bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Schema: "ops", Table: "devices"},
		RenameTo:     "device",
		AddColumns:   []bo.SqlColumnDefinition{{Name: "owner", Type: "text", Nullable: true}},
		DropColumns:  []string{"legacy"},
		AlterColumns: []bo.SqlAlterColumn{{Name: "name", Type: "varchar(128)", Nullable: &nullable}},
	})
	if err != nil {
		t.Fatalf("buildAlterSqlTable: %v", err)
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
}

func TestBuildAlterSqlTableRequiresChange(t *testing.T) {
	if _, err := buildAlterSqlTable(bo.AlterSqlTableInput{SqlTableRef: bo.SqlTableRef{Table: "devices"}}); !merr.IsInvalid(err) {
		t.Fatalf("empty alter err = %v", err)
	}
	if _, err := buildAlterSqlTable(bo.AlterSqlTableInput{
		SqlTableRef:  bo.SqlTableRef{Table: "devices"},
		AlterColumns: []bo.SqlAlterColumn{{Name: "name"}},
	}); !merr.IsInvalid(err) {
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

func TestSqlDSN(t *testing.T) {
	dsn, err := sqlDSN(bo.Datasource{
		Endpoint: "postgres://db.moon.com:5432/app",
		Username: "moon",
		Password: "p@ss/w0rd",
	}, "analytics")
	if err != nil {
		t.Fatalf("sqlDSN: %v", err)
	}
	for _, want := range []string{"postgres://", "db.moon.com:5432/analytics", "application_name=alioth"} {
		if !strings.Contains(dsn, want) {
			t.Fatalf("dsn missing %q: %s", want, dsn)
		}
	}
	if strings.Contains(dsn, "p@ss/w0rd") {
		t.Fatalf("dsn should escape the password: %s", dsn)
	}

	if _, err := sqlDSN(bo.Datasource{Endpoint: "http://db.moon.com"}, ""); !merr.IsInvalid(err) {
		t.Fatalf("http endpoint err = %v", err)
	}
	if _, err := sqlDSN(bo.Datasource{Endpoint: "postgres://db.moon.com/app"}, "bad/name"); !merr.IsInvalid(err) {
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

func TestBuildSqlTableDataStatement(t *testing.T) {
	got := buildSqlTableDataStatement("public", "devices", []string{"id", "region"}, 101, 200)
	want := `SELECT * FROM "public"."devices" ORDER BY "id", "region" LIMIT 101 OFFSET 200`
	if got != want {
		t.Fatalf("statement = %s, want %s", got, want)
	}
	// 没有主键时按物理行号排序，保证翻页有确定的顺序。
	got = buildSqlTableDataStatement("ops", "events", nil, 100, 0)
	want = `SELECT * FROM "ops"."events" ORDER BY ctid LIMIT 100 OFFSET 0`
	if got != want {
		t.Fatalf("statement = %s, want %s", got, want)
	}
}
