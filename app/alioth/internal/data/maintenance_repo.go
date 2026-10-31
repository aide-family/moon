package data

import (
	"bufio"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// archiveFormatVersion 是迁移包格式版本。导入只接受完全相同的版本，
	// 格式升级时在迁移包与旧版本之间显式做转换，不在导入路径里猜。
	archiveFormatVersion = 1
	// archiveBatchSize 是一次 INSERT 覆盖的行数。
	archiveBatchSize = 500
	// archiveScannerLimit 是单行 JSON 的上限：集群 kubeconfig 等密文可能很长。
	archiveScannerLimit = 16 << 20
)

// archivedTables 是迁移包覆盖的配置与身份表，顺序即导出顺序。
// 运行数据（实时告警、任务队列、验证码、进程心跳）不进迁移包，见 runtimeTables。
var archivedTables = []string{
	// 身份与租户。
	"users",
	"user_oauths",
	"namespaces",
	"members",
	"member_invites",
	"member_roles",
	"roles",
	"role_permissions",
	"permissions",
	"system_messages",
	"user_alert_pages",
	// 套餐与订单。
	"feature_packages",
	"feature_package_items",
	"namespace_features",
	"namespace_orders",
	// 告警配置。
	"datasources",
	"alert_dictionaries",
	"alert_rules",
	"alert_rule_groups",
	"alert_rule_levels",
	"alert_pages",
	"alert_event_columns",
	"probe_target_groups",
	"probe_targets",
	// 通知配置。
	"notification_channels",
	"notification_templates",
	"notification_presets",
	"notification_groups",
	"notification_on_calls",
	"notification_on_call_groups",
	"notification_subscriptions",
	"notification_subscription_groups",
	// 边缘。
	"clusters",
	"agents",
	"ssh_commands",
	"ssh_command_audits",
}

// runtimeTables 是明确不迁的运行数据。表清单与 ent 的 migrate.Tables 由单测对照，
// 新增表必须二选一，不能悄悄漏掉。
var runtimeTables = []string{
	"alert_eval_batches",
	"alert_events",
	"alert_event_firings",
	"alert_history_events",
	"captchas",
	"email_login_codes",
	"message_logs",
	"message_log_shards",
	"notification_tasks",
	"probe_results",
	"worker_instances",
}

var (
	archivedTableIndex = func() map[string]int {
		index := make(map[string]int, len(archivedTables))
		for position, name := range archivedTables {
			index[name] = position
		}
		return index
	}()
	// identifierPattern 限制迁移包里的表名与列名，拼接 SQL 前必须命中。
	identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

type maintenanceRepo struct {
	deps *Dependencies
	log  *slog.Logger
}

// NewMaintenanceRepo 提供平台升级与迁移的数据操作。
func NewMaintenanceRepo(deps *Dependencies, logger *slog.Logger) repository.MaintenanceRepo {
	return &maintenanceRepo{deps: deps, log: logger}
}

// archiveManifest 是迁移包的第一行。列清单按导出时的表结构记录，
// 导入时用它决定插哪些列，升级后新增的列走目标库默认值。
type archiveManifest struct {
	FormatVersion   int32                  `json:"formatVersion"`
	AppVersion      string                 `json:"appVersion"`
	CreatedAt       string                 `json:"createdAt"`
	EncryptKeyCheck string                 `json:"encryptKeyCheck"`
	Tables          []archiveManifestTable `json:"tables"`
}

type archiveManifestTable struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

// archiveStats 是迁移包的最后一行，记录每张表实际导出的行数。
// 导入必须读到它；缺这一行说明文件被截断。
type archiveStats struct {
	Tables []archiveTableStat `json:"tables"`
}

type archiveTableStat struct {
	Name string `json:"name"`
	Rows int64  `json:"rows"`
}

// archiveEnvelope 是迁移包的一行：manifest、row、summary 三种 kind。
type archiveEnvelope struct {
	Kind     string           `json:"kind"`
	Table    string           `json:"table,omitempty"`
	Row      json.RawMessage  `json:"row,omitempty"`
	Manifest *archiveManifest `json:"manifest,omitempty"`
	Summary  *archiveStats    `json:"summary,omitempty"`
}

// Status 读取数据库版本、配置表行数与运行表行数估算。
// 配置表用 count(*) 精确统计；运行表可能很大，用 pg_class.reltuples 估算。
func (repo *maintenanceRepo) Status(ctx context.Context) (*bo.MaintenanceDatabase, error) {
	version, err := repo.postgresVersion(ctx)
	if err != nil {
		return nil, err
	}
	tables := make([]bo.MaintenanceTable, 0, len(archivedTables)+len(runtimeTables))
	for _, name := range archivedTables {
		exists, err := repo.tableExists(ctx, name)
		if err != nil {
			return nil, err
		}
		var rows int64
		if exists {
			if rows, err = repo.countRows(ctx, name); err != nil {
				return nil, err
			}
		}
		tables = append(tables, bo.MaintenanceTable{Name: name, Archived: true, Rows: rows, Exists: exists})
	}
	for _, name := range runtimeTables {
		exists, err := repo.tableExists(ctx, name)
		if err != nil {
			return nil, err
		}
		var rows int64
		if exists {
			if rows, err = repo.estimateRows(ctx, name); err != nil {
				return nil, err
			}
		}
		tables = append(tables, bo.MaintenanceTable{Name: name, Archived: false, Rows: rows, Exists: exists})
	}
	return &bo.MaintenanceDatabase{
		PostgreSQLVersion: version,
		EncryptKeyCheck:   repo.deps.encryptKeyCheck,
		Tables:            tables,
	}, nil
}

// MigrateSchema 执行 ent 建表、seed 与周分区，与 alioth migrate 同一套代码。
func (repo *maintenanceRepo) MigrateSchema(ctx context.Context) (bo.SchemaMigration, error) {
	started := time.Now()
	if err := ApplySchema(ctx, repo.deps.client, repo.deps.db); err != nil {
		return bo.SchemaMigration{}, databaseErr(repo.log, err)
	}
	if err := seed(ctx, repo.deps.client, repo.log); err != nil {
		return bo.SchemaMigration{}, err
	}
	return bo.SchemaMigration{
		TableCount: int32(len(archivedTables) + len(runtimeTables)),
		Duration:   time.Since(started),
	}, nil
}

// ExportArchive 以 gzip 流写出迁移包：
// 第一行 manifest，随后逐表写 row，最后一行 summary。
func (repo *maintenanceRepo) ExportArchive(ctx context.Context, w io.Writer, appVersion string) (bo.ArchiveExport, error) {
	manifest, err := repo.buildArchiveManifest(ctx, appVersion)
	if err != nil {
		return bo.ArchiveExport{}, err
	}
	gz := gzip.NewWriter(w)
	buffered := bufio.NewWriterSize(gz, 64*1024)
	if err := writeArchiveLine(buffered, archiveEnvelope{Kind: "manifest", Manifest: manifest}); err != nil {
		return bo.ArchiveExport{}, archiveWriteErr()
	}
	stats := make([]archiveTableStat, 0, len(manifest.Tables))
	var exported bo.ArchiveExport
	for _, table := range manifest.Tables {
		rows, err := repo.exportArchiveTable(ctx, buffered, table)
		if err != nil {
			return bo.ArchiveExport{}, err
		}
		stats = append(stats, archiveTableStat{Name: table.Name, Rows: rows})
		exported.Tables++
		exported.Rows += rows
	}
	if err := writeArchiveLine(buffered, archiveEnvelope{Kind: "summary", Summary: &archiveStats{Tables: stats}}); err != nil {
		return bo.ArchiveExport{}, archiveWriteErr()
	}
	if err := buffered.Flush(); err != nil {
		return bo.ArchiveExport{}, archiveWriteErr()
	}
	if err := gz.Close(); err != nil {
		return bo.ArchiveExport{}, archiveWriteErr()
	}
	return exported, nil
}

// ImportArchive 校验并覆盖导入迁移包。整笔导入在一个事务里完成，
// 成功后提交事务并补跑 seed；任何一步失败都回滚，目标库保持原样。
func (repo *maintenanceRepo) ImportArchive(ctx context.Context, r io.Reader) (bo.ArchiveImport, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return bo.ArchiveImport{}, merr.Client("invalid archive")
	}
	defer func() { _ = reader.Close() }()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), archiveScannerLimit)

	manifest, err := readArchiveManifest(scanner)
	if err != nil {
		return bo.ArchiveImport{}, err
	}
	if manifest.FormatVersion != archiveFormatVersion {
		return bo.ArchiveImport{}, merr.Client("archive format is not supported")
	}
	columns, err := repo.validateArchiveManifest(ctx, manifest)
	if err != nil {
		return bo.ArchiveImport{}, err
	}
	// 第一步先确认目标库能接收迁移包，再开始改动数据。
	if err := repo.ensureMigrationTables(ctx); err != nil {
		return bo.ArchiveImport{}, err
	}

	tx, err := repo.deps.db.BeginTx(ctx, nil)
	if err != nil {
		return bo.ArchiveImport{}, databaseErr(repo.log, err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := truncateMigrationTarget(ctx, tx); err != nil {
		return bo.ArchiveImport{}, databaseErr(repo.log, err)
	}

	pending := archiveBatch{}
	counts := make(map[string]int64, len(manifest.Tables))
	var summary *archiveStats
	finished := false
	for scanner.Scan() {
		var line archiveEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return bo.ArchiveImport{}, merr.Client("invalid archive")
		}
		switch line.Kind {
		case "row":
			tableColumns, ok := columns[line.Table]
			if !ok || len(line.Row) == 0 {
				return bo.ArchiveImport{}, merr.Client("invalid archive")
			}
			if pending.table != line.Table {
				if err := flushArchiveBatch(ctx, tx, pending); err != nil {
					return bo.ArchiveImport{}, err
				}
				pending = archiveBatch{table: line.Table, columns: tableColumns}
			}
			pending.rows = append(pending.rows, string(line.Row))
			counts[line.Table]++
			if len(pending.rows) >= archiveBatchSize {
				if err := flushArchiveBatch(ctx, tx, pending); err != nil {
					return bo.ArchiveImport{}, err
				}
				pending.rows = pending.rows[:0]
			}
		case "summary":
			if line.Summary == nil {
				return bo.ArchiveImport{}, merr.Client("invalid archive")
			}
			summary = line.Summary
			finished = true
		default:
			return bo.ArchiveImport{}, merr.Client("invalid archive")
		}
		if finished {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return bo.ArchiveImport{}, merr.Client("invalid archive")
	}
	if summary == nil {
		return bo.ArchiveImport{}, merr.Client("archive is incomplete")
	}
	if err := flushArchiveBatch(ctx, tx, pending); err != nil {
		return bo.ArchiveImport{}, err
	}
	if err := verifyArchiveCounts(summary, counts, manifest.Tables); err != nil {
		return bo.ArchiveImport{}, err
	}
	if err := resetArchiveSequences(ctx, tx); err != nil {
		return bo.ArchiveImport{}, databaseErr(repo.log, err)
	}
	if err := tx.Commit(); err != nil {
		return bo.ArchiveImport{}, databaseErr(repo.log, err)
	}
	imported := bo.ArchiveImport{Tables: int32(len(summary.Tables)), SourceVersion: manifest.AppVersion}
	for _, item := range summary.Tables {
		imported.Rows += item.Rows
	}
	// 导入完成后补跑 seed：新版本新增的权限、预置模板、基础套餐和租户快照
	// 都要在迁移包里没有的情况下补齐。seed 幂等，失败时可以再点一次数据库迁移。
	if err := seed(ctx, repo.deps.client, repo.log); err != nil {
		repo.log.Error(fmt.Sprintf("seed after import: %v", err))
		return imported, merr.Server("archive imported but seeding failed; run the database migration to finish")
	}
	return imported, nil
}

// buildArchiveManifest 读取每张配置表的列清单。缺表说明目标库还没迁移。
func (repo *maintenanceRepo) buildArchiveManifest(ctx context.Context, appVersion string) (*archiveManifest, error) {
	manifest := &archiveManifest{
		FormatVersion:   archiveFormatVersion,
		AppVersion:      appVersion,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		EncryptKeyCheck: repo.deps.encryptKeyCheck,
		Tables:          make([]archiveManifestTable, 0, len(archivedTables)),
	}
	for _, name := range archivedTables {
		columns, err := repo.tableColumns(ctx, name)
		if err != nil {
			return nil, err
		}
		if len(columns) == 0 {
			return nil, merr.Client("database schema is not ready")
		}
		manifest.Tables = append(manifest.Tables, archiveManifestTable{Name: name, Columns: columns})
	}
	return manifest, nil
}

// validateArchiveManifest 校验迁移包的版本、密钥与列集合，返回按表分组的列清单。
func (repo *maintenanceRepo) validateArchiveManifest(ctx context.Context, manifest *archiveManifest) (map[string][]string, error) {
	if manifest.EncryptKeyCheck != repo.deps.encryptKeyCheck {
		return nil, merr.Client("encrypt key does not match the archive")
	}
	if len(manifest.Tables) == 0 {
		return nil, merr.Client("invalid archive")
	}
	columns := make(map[string][]string, len(manifest.Tables))
	for _, table := range manifest.Tables {
		if _, ok := archivedTableIndex[table.Name]; !ok {
			return nil, merr.Client("archive contains an unsupported table")
		}
		if _, seen := columns[table.Name]; seen {
			return nil, merr.Client("invalid archive")
		}
		if len(table.Columns) == 0 {
			return nil, merr.Client("invalid archive")
		}
		current, err := repo.tableColumns(ctx, table.Name)
		if err != nil {
			return nil, err
		}
		currentSet := make(map[string]struct{}, len(current))
		for _, name := range current {
			currentSet[name] = struct{}{}
		}
		for _, name := range table.Columns {
			if !identifierPattern.MatchString(name) {
				return nil, merr.Client("invalid archive")
			}
			if _, ok := currentSet[name]; !ok {
				return nil, merr.Client("archive column is not supported by the target database")
			}
		}
		columns[table.Name] = table.Columns
	}
	return columns, nil
}

// ensureMigrationTables 确认导入需要读写的表都已经建好，导入前就能给出明确提示。
func (repo *maintenanceRepo) ensureMigrationTables(ctx context.Context) error {
	for _, name := range append(append([]string{}, archivedTables...), runtimeTables...) {
		exists, err := repo.tableExists(ctx, name)
		if err != nil {
			return err
		}
		if !exists {
			return merr.Client("database schema is not ready")
		}
	}
	return nil
}

// exportArchiveTable 用 row_to_json 逐行写出，不把整表读进内存。
func (repo *maintenanceRepo) exportArchiveTable(ctx context.Context, w *bufio.Writer, table archiveManifestTable) (int64, error) {
	query := fmt.Sprintf(`SELECT row_to_json(t)::text FROM (SELECT * FROM %s) t`, quoteIdentifier(table.Name))
	rows, err := repo.deps.db.QueryContext(ctx, query)
	if err != nil {
		return 0, databaseErr(repo.log, err)
	}
	defer func() { _ = rows.Close() }()
	var count int64
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return 0, databaseErr(repo.log, err)
		}
		if err := writeArchiveLine(w, archiveEnvelope{Kind: "row", Table: table.Name, Row: json.RawMessage(raw)}); err != nil {
			return 0, archiveWriteErr()
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return 0, databaseErr(repo.log, err)
	}
	return count, nil
}

// truncateMigrationTarget 清空迁移包覆盖的配置表和全部运行数据。
// 运行数据不进迁移包，导入后保留旧环境的队列与实时事件只会留下孤立引用，
// 所以这里显式清空；历史分库在独立库里，不在这条语句的范围内。
func truncateMigrationTarget(ctx context.Context, tx *sql.Tx) error {
	names := make([]string, 0, len(archivedTables)+len(runtimeTables))
	for _, name := range archivedTables {
		names = append(names, quoteIdentifier(name))
	}
	for _, name := range runtimeTables {
		names = append(names, quoteIdentifier(name))
	}
	query := `TRUNCATE TABLE ` + strings.Join(names, ", ") + ` RESTART IDENTITY CASCADE`
	_, err := tx.ExecContext(ctx, query)
	return err
}

// resetArchiveSequences 把每张配置表的 id 序列推到当前最大值之后。
// 导入带着原 id 写入，不重置会让迁移后新建记录撞主键。
func resetArchiveSequences(ctx context.Context, tx *sql.Tx) error {
	for _, name := range archivedTables {
		var sequence sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, name).Scan(&sequence); err != nil {
			return err
		}
		if !sequence.Valid || sequence.String == "" {
			continue
		}
		var next int64
		query := fmt.Sprintf(`SELECT COALESCE(MAX(id), 0) + 1 FROM %s`, quoteIdentifier(name))
		if err := tx.QueryRowContext(ctx, query).Scan(&next); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `SELECT setval($1::regclass, $2::bigint, false)`, sequence.String, next); err != nil {
			return err
		}
	}
	return nil
}

type archiveBatch struct {
	table   string
	columns []string
	rows    []string
}

// flushArchiveBatch 把一批行用 json_populate_recordset 转成表行插入。
func flushArchiveBatch(ctx context.Context, tx *sql.Tx, batch archiveBatch) error {
	if batch.table == "" || len(batch.rows) == 0 {
		return nil
	}
	quotedColumns := make([]string, 0, len(batch.columns))
	for _, name := range batch.columns {
		if !identifierPattern.MatchString(name) {
			return merr.Client("invalid archive")
		}
		quotedColumns = append(quotedColumns, quoteIdentifier(name))
	}
	columnList := strings.Join(quotedColumns, ", ")
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) SELECT %s FROM json_populate_recordset(NULL::%s, $1::json)`,
		quoteIdentifier(batch.table), columnList, columnList, quoteIdentifier(batch.table),
	)
	payload := "[" + strings.Join(batch.rows, ",") + "]"
	if _, err := tx.ExecContext(ctx, query, payload); err != nil {
		return merr.Client("archive data does not match the target database")
	}
	return nil
}

// verifyArchiveCounts 用末尾的 summary 核对每张表实际读到的行数。
func verifyArchiveCounts(summary *archiveStats, counts map[string]int64, tables []archiveManifestTable) error {
	if len(summary.Tables) == 0 {
		return merr.Client("invalid archive")
	}
	seen := make(map[string]struct{}, len(summary.Tables))
	for _, item := range summary.Tables {
		if _, ok := archivedTableIndex[item.Name]; !ok {
			return merr.Client("archive contains an unsupported table")
		}
		if _, ok := seen[item.Name]; ok {
			return merr.Client("invalid archive")
		}
		seen[item.Name] = struct{}{}
		if counts[item.Name] != item.Rows {
			return merr.Client("archive is incomplete")
		}
	}
	// manifest 里列出的表必须都在 summary 里出现，否则说明行数与表清单对不上。
	for _, table := range tables {
		if _, ok := seen[table.Name]; !ok {
			return merr.Client("archive is incomplete")
		}
	}
	for name := range counts {
		if _, ok := seen[name]; !ok {
			return merr.Client("invalid archive")
		}
	}
	return nil
}

// readArchiveManifest 读取并解析第一行；第一行必须是 manifest。
func readArchiveManifest(scanner *bufio.Scanner) (*archiveManifest, error) {
	if !scanner.Scan() {
		return nil, merr.Client("invalid archive")
	}
	var line archiveEnvelope
	if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
		return nil, merr.Client("invalid archive")
	}
	if line.Kind != "manifest" || line.Manifest == nil {
		return nil, merr.Client("invalid archive")
	}
	return line.Manifest, nil
}

func (repo *maintenanceRepo) postgresVersion(ctx context.Context) (string, error) {
	var version string
	if err := repo.deps.db.QueryRowContext(ctx, `SELECT current_setting('server_version')`).Scan(&version); err != nil {
		return "", databaseErr(repo.log, err)
	}
	return version, nil
}

func (repo *maintenanceRepo) tableExists(ctx context.Context, name string) (bool, error) {
	var found sql.NullString
	if err := repo.deps.db.QueryRowContext(ctx, `SELECT to_regclass($1)::text`, name).Scan(&found); err != nil {
		return false, databaseErr(repo.log, err)
	}
	return found.Valid && found.String != "", nil
}

func (repo *maintenanceRepo) countRows(ctx context.Context, name string) (int64, error) {
	var count int64
	query := fmt.Sprintf(`SELECT count(*) FROM %s`, quoteIdentifier(name))
	if err := repo.deps.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, databaseErr(repo.log, err)
	}
	return count, nil
}

func (repo *maintenanceRepo) estimateRows(ctx context.Context, name string) (int64, error) {
	var count int64
	if err := repo.deps.db.QueryRowContext(ctx,
		`SELECT GREATEST(COALESCE((SELECT reltuples::bigint FROM pg_class WHERE oid = to_regclass($1)), 0), 0)`, name,
	).Scan(&count); err != nil {
		return 0, databaseErr(repo.log, err)
	}
	return count, nil
}

// tableColumns 返回表的列名，顺序与物理列顺序一致。
func (repo *maintenanceRepo) tableColumns(ctx context.Context, name string) ([]string, error) {
	rows, err := repo.deps.db.QueryContext(ctx, `
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1
		ORDER BY ordinal_position`, name)
	if err != nil {
		return nil, databaseErr(repo.log, err)
	}
	defer func() { _ = rows.Close() }()
	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, databaseErr(repo.log, err)
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseErr(repo.log, err)
	}
	return columns, nil
}

func writeArchiveLine(w *bufio.Writer, line archiveEnvelope) error {
	payload, err := json.Marshal(line)
	if err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return w.WriteByte('\n')
}

// archiveWriteErr 把写 tar/gzip 的 IO 错误收口成服务端错误，具体原因只在日志里出现。
func archiveWriteErr() error {
	return merr.Server("archive write failed")
}
