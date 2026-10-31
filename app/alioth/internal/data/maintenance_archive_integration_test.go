package data

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// 迁移包只有真库才能验证：row_to_json / json_populate_recordset 的列顺序、
// 时间与小数字段、序列重置都依赖 PostgreSQL。测试会覆盖测试库的配置表，
// 所以只接受 ALIOTH_TEST_DSN，不允许指向开发库 alioth。
func TestArchiveExportImportAgainstRealDatabase(t *testing.T) {
	dsn := testDSN(t)
	db := openTestDB(t, dsn)
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	deps := &Dependencies{db: db, client: client, encryptKeyCheck: "archive-test-key"}
	repo := NewMaintenanceRepo(deps, slog.Default())
	ctx := context.Background()

	if _, err := repo.MigrateSchema(ctx); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}
	if _, err := repo.MigrateSchema(ctx); err != nil {
		t.Fatalf("migrate schema is not repeatable: %v", err)
	}

	suffix := time.Now().UnixNano()
	uid := fmt.Sprintf("migration-user-%d", suffix)
	email := fmt.Sprintf("migration-%d@example.com", suffix)
	created, err := client.User.Create().SetUID(uid).SetEmail(email).SetName("migration").Save(ctx)
	if err != nil {
		t.Fatalf("create migration marker user: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("marker user has no id")
	}
	// 运行数据不进迁移包，导入时必须被清空；这里插一条验证码作为运行数据标记。
	if _, err := db.ExecContext(ctx,
		`INSERT INTO captchas (created_at, updated_at, uid, answer_hash, expires_at) VALUES (now(), now(), $1, 'hash', now() + interval '5 minutes')`,
		fmt.Sprintf("migration-captcha-%d", suffix)); err != nil {
		t.Fatalf("insert runtime marker: %v", err)
	}
	nextUID := fmt.Sprintf("migration-user-next-%d", suffix)
	t.Cleanup(func() {
		_, _ = client.User.Delete().Where(user.UIDIn(uid, nextUID)).Exec(context.Background())
	})

	before, err := repo.Status(ctx)
	if err != nil {
		t.Fatalf("status before export: %v", err)
	}
	if !databaseSchemaReady(before) {
		t.Fatal("schema should be ready after migrate")
	}

	var archive bytes.Buffer
	exported, err := repo.ExportArchive(ctx, &archive, "test-version")
	if err != nil {
		t.Fatalf("export archive: %v", err)
	}
	manifest, summary := inspectArchive(t, archive.Bytes())
	if manifest.FormatVersion != archiveFormatVersion {
		t.Fatalf("manifest format = %d, want %d", manifest.FormatVersion, archiveFormatVersion)
	}
	if manifest.EncryptKeyCheck != deps.encryptKeyCheck {
		t.Fatalf("manifest key check = %q, want %q", manifest.EncryptKeyCheck, deps.encryptKeyCheck)
	}
	if len(manifest.Tables) != len(archivedTables) {
		t.Fatalf("manifest tables = %d, want %d", len(manifest.Tables), len(archivedTables))
	}
	var rows int64
	for _, item := range summary.Tables {
		rows += item.Rows
	}
	if exported.Tables != int32(len(archivedTables)) || exported.Rows != rows {
		t.Fatalf("exported = %+v, want %d tables and %d rows", exported, len(archivedTables), rows)
	}

	imported, err := repo.ImportArchive(ctx, bytes.NewReader(archive.Bytes()))
	if err != nil {
		t.Fatalf("import archive: %v", err)
	}
	if imported.SourceVersion != "test-version" {
		t.Fatalf("imported source version = %q, want test-version", imported.SourceVersion)
	}
	if imported.Tables != exported.Tables || imported.Rows != exported.Rows {
		t.Fatalf("imported = %+v, want %+v", imported, exported)
	}
	after, err := repo.Status(ctx)
	if err != nil {
		t.Fatalf("status after import: %v", err)
	}
	assertArchivedRowsEqual(t, before, after)

	if alive, err := client.User.Query().Where(user.UIDEQ(uid)).Exist(ctx); err != nil {
		t.Fatalf("query marker user: %v", err)
	} else if !alive {
		t.Fatal("marker user disappeared after import")
	}
	var runtimeRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM captchas`).Scan(&runtimeRows); err != nil {
		t.Fatalf("count runtime rows: %v", err)
	}
	if runtimeRows != 0 {
		t.Fatalf("runtime rows = %d after import, want 0", runtimeRows)
	}

	// 导入带着原 id 写入，序列必须重置；否则下一次自增插入会撞主键。
	if _, err := client.User.Create().
		SetUID(nextUID).
		SetEmail(fmt.Sprintf("migration-next-%d@example.com", suffix)).
		SetName("migration next").
		Save(ctx); err != nil {
		t.Fatalf("insert after import: %v", err)
	}

	// 密钥不一致必须在清空数据之前拒绝。
	wrongKey := &maintenanceRepo{
		deps: &Dependencies{db: db, client: client, encryptKeyCheck: "another-key"},
		log:  slog.Default(),
	}
	if _, err := wrongKey.ImportArchive(ctx, bytes.NewReader(archive.Bytes())); err == nil {
		t.Fatal("import with a different encrypt key must fail")
	}
	// 截断的文件也要被拒绝，且目标库保持原样。
	truncated := archive.Bytes()[:len(archive.Bytes())/2]
	if _, err := repo.ImportArchive(ctx, bytes.NewReader(truncated)); err == nil {
		t.Fatal("truncated archive must fail")
	}
	if alive, err := client.User.Query().Where(user.UIDEQ(uid)).Exist(ctx); err != nil {
		t.Fatalf("query marker user after failed imports: %v", err)
	} else if !alive {
		t.Fatal("failed imports must not change the target database")
	}
}

// inspectArchive 解析迁移包的 manifest 与 summary，并确认行数落在 summary 里。
func inspectArchive(t *testing.T, raw []byte) (*archiveManifest, *archiveStats) {
	t.Helper()
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open gzip: %v", err)
	}
	defer func() { _ = reader.Close() }()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), archiveScannerLimit)
	var manifest *archiveManifest
	var summary *archiveStats
	for scanner.Scan() {
		var line archiveEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("decode archive line: %v", err)
		}
		switch line.Kind {
		case "manifest":
			manifest = line.Manifest
		case "summary":
			summary = line.Summary
		case "row":
		default:
			t.Fatalf("unexpected archive line kind %q", line.Kind)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan archive: %v", err)
	}
	if manifest == nil || summary == nil {
		t.Fatal("archive is missing manifest or summary")
	}
	return manifest, summary
}

func databaseSchemaReady(database *bo.MaintenanceDatabase) bool {
	for _, table := range database.Tables {
		if table.Archived && !table.Exists {
			return false
		}
	}
	return true
}

func assertArchivedRowsEqual(t *testing.T, before, after *bo.MaintenanceDatabase) {
	t.Helper()
	previous := make(map[string]int64, len(before.Tables))
	for _, table := range before.Tables {
		if table.Archived {
			previous[table.Name] = table.Rows
		}
	}
	for _, table := range after.Tables {
		if !table.Archived {
			continue
		}
		if got := previous[table.Name]; got != table.Rows {
			t.Fatalf("table %s rows = %d after import, want %d", table.Name, table.Rows, got)
		}
	}
}
