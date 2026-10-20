package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	entschema "entgo.io/ent/dialect/sql/schema"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/pkg/merr"
)

var partitionedTables = []string{"alert_events", "message_logs"}

// ApplySchema 创建表，并把空的告警事件和消息日志转成按 created_at 的周分区表。
// 已经分区时跳过这些表，避免生成器把主键改回单列 id。
func ApplySchema(ctx context.Context, client *ent.Client, db *sql.DB) error {
	if err := renameAlertDictionaries(ctx, db); err != nil {
		return err
	}
	if err := prepareAlertEvalBatches(ctx, db); err != nil {
		return err
	}
	if err := prepareAlertEventIntervene(ctx, db); err != nil {
		return err
	}
	opts := []entschema.MigrateOption{}
	for _, name := range partitionedTables {
		kind, err := relationKind(ctx, db, name)
		if err != nil {
			return err
		}
		if kind == "p" {
			opts = append(opts, entschema.WithHooks(skipTable(name)))
		}
	}
	if err := client.Schema.Create(ctx, opts...); err != nil {
		return err
	}
	for _, name := range partitionedTables {
		if err := partitionByCreatedAt(ctx, db, name); err != nil {
			return err
		}
		if err := ensureWeekPartitions(ctx, db, name, time.Now().UTC()); err != nil {
			return err
		}
		if err := ensurePartitionIndexes(ctx, db, name); err != nil {
			return err
		}
	}
	return nil
}

// partitionIndexes 是分区父表必须保留的查询索引。
// CREATE TABLE ... (LIKE ...) 不会复制索引，转换后必须显式重建；这里用 IF NOT EXISTS 保证重复迁移安全。
var partitionIndexes = map[string][]string{
	"alert_events": {
		`CREATE INDEX IF NOT EXISTS alert_events_ns_status_created ON alert_events (namespace_uid, status, created_at)`,
		`CREATE INDEX IF NOT EXISTS alert_events_rule_uid ON alert_events (rule_uid)`,
		`CREATE INDEX IF NOT EXISTS alert_events_fingerprint ON alert_events (fingerprint)`,
	},
	"message_logs": {
		`CREATE INDEX IF NOT EXISTS message_logs_ns_created ON message_logs (namespace_uid, created_at)`,
		`CREATE INDEX IF NOT EXISTS message_logs_task_uid ON message_logs (task_uid)`,
	},
}

func ensurePartitionIndexes(ctx context.Context, db *sql.DB, name string) error {
	kind, err := relationKind(ctx, db, name)
	if err != nil {
		return err
	}
	if kind != "p" {
		return nil
	}
	for _, statement := range partitionIndexes[name] {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("index %s: %v", name, err))
			return merr.Server("request failed")
		}
	}
	return nil
}

// renameAlertDictionaries 把旧的 alert_levels 表和索引改成告警字典的名字。
// ent 的 Schema.Create 不会改已有表名，不改的话会另建一张空表。
func renameAlertDictionaries(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_levels")
	if err != nil || kind != "r" {
		return err
	}
	statements := []string{
		`ALTER TABLE alert_levels RENAME TO alert_dictionaries`,
		`ALTER INDEX IF EXISTS alertlevel_uid RENAME TO alertdictionary_uid`,
		`ALTER INDEX IF EXISTS alertlevel_namespace_uid_name RENAME TO alertdictionary_namespace_uid_name`,
		`ALTER INDEX IF EXISTS alertlevel_namespace_uid RENAME TO alertdictionary_namespace_uid`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("schema: %v", err))
			return merr.Server("request failed")
		}
	}
	return nil
}

// prepareAlertEvalBatches 给旧表补上 datasource_uid。
// 这列是后加的非空字段，表里已有行时直接 ALTER 会因空值失败。
// 旧的 (rule_uid, eval_slot) 唯一索引也要去掉，改由 (rule_uid, datasource_uid, eval_slot) 约束。
func prepareAlertEvalBatches(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_eval_batches")
	if err != nil || kind != "r" {
		return err
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'alert_eval_batches' AND column_name = 'datasource_uid'
	)`).Scan(&exists); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	if !exists {
		if _, err := db.ExecContext(ctx, `ALTER TABLE alert_eval_batches ADD COLUMN datasource_uid character varying NOT NULL DEFAULT ''`); err != nil {
			slog.Default().Error(fmt.Sprintf("schema: %v", err))
			return merr.Server("request failed")
		}
		if _, err := db.ExecContext(ctx, `ALTER TABLE alert_eval_batches ALTER COLUMN datasource_uid DROP DEFAULT`); err != nil {
			slog.Default().Error(fmt.Sprintf("schema: %v", err))
			return merr.Server("request failed")
		}
	}
	if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS alertevalbatch_rule_uid_eval_slot`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

// prepareAlertEventIntervene 给已分区的告警事件补上介入时间。
// 分区表会被迁移跳过，父表加列会落到各个分区。
func prepareAlertEventIntervene(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_events")
	if err != nil || kind == "" {
		return err
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE alert_events ADD COLUMN IF NOT EXISTS intervened_at timestamptz NULL`); err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return merr.Server("request failed")
	}
	return nil
}

func skipTable(name string) entschema.Hook {
	return func(next entschema.Creator) entschema.Creator {
		return entschema.CreateFunc(func(ctx context.Context, tables ...*entschema.Table) error {
			kept := make([]*entschema.Table, 0, len(tables))
			for _, table := range tables {
				if table.Name == name {
					continue
				}
				kept = append(kept, table)
			}
			return next.Create(ctx, kept...)
		})
	}
}

func partitionByCreatedAt(ctx context.Context, db *sql.DB, name string) error {
	kind, err := relationKind(ctx, db, name)
	if err != nil {
		return err
	}
	if kind == "p" || kind == "" {
		return nil
	}
	if kind != "r" {
		return merr.Server(name + " cannot be partitioned")
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+name).Scan(&count); err != nil {
		slog.Default().Error(fmt.Sprintf("count %s: %v", name, err))
		return merr.Server("request failed")
	}
	if count > 0 {
		return merr.Server(name + " is not empty")
	}
	heap := name + "_heap"
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return merr.Server("request failed")
	}
	defer func() { _ = tx.Rollback() }()
	// 旧表改名后，约束、索引和 identity 序列的名字不会跟着变。
	// 新分区表要复用 table_pkey / table_id_seq 这些默认名，所以先把旧名字释放掉。
	statements := []string{
		fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, name, heap),
		fmt.Sprintf(`ALTER SEQUENCE IF EXISTS %s_id_seq RENAME TO %s_id_seq`, name, heap),
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("partition %s: %v", name, err))
			return merr.Server("request failed")
		}
	}
	if err := dropPrimaryKey(ctx, tx, heap); err != nil {
		return err
	}
	if err := dropIndexes(ctx, tx, heap); err != nil {
		return err
	}
	create := []string{
		fmt.Sprintf(`CREATE TABLE %s (LIKE %s INCLUDING DEFAULTS INCLUDING IDENTITY) PARTITION BY RANGE (created_at)`, name, heap),
		fmt.Sprintf(`ALTER TABLE %s ADD PRIMARY KEY (created_at, id)`, name),
		fmt.Sprintf(`CREATE UNIQUE INDEX %s_uid_created_at ON %s (uid, created_at)`, name, name),
	}
	for _, statement := range create {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("partition %s: %v", name, err))
			return merr.Server("request failed")
		}
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DROP TABLE %s`, heap)); err != nil {
		slog.Default().Error(fmt.Sprintf("drop %s heap: %v", name, err))
		return merr.Server("request failed")
	}
	return tx.Commit()
}

// dropPrimaryKey 删除临时表的主键约束，释放默认的 <table>_pkey 索引名。
func dropPrimaryKey(ctx context.Context, tx *sql.Tx, table string) error {
	var name string
	err := tx.QueryRowContext(ctx, `SELECT conname FROM pg_constraint WHERE conrelid = $1::regclass AND contype = 'p'`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		slog.Default().Error(fmt.Sprintf("primary key %s: %v", table, err))
		return merr.Server("request failed")
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT %q`, table, name)); err != nil {
		slog.Default().Error(fmt.Sprintf("drop primary key %s: %v", table, err))
		return merr.Server("request failed")
	}
	return nil
}

// dropIndexes 删除临时表上的普通索引。主键约束已经先删掉，剩余索引随 DROP TABLE 丢弃，这里只是释放索引名。
func dropIndexes(ctx context.Context, tx *sql.Tx, table string) error {
	rows, err := tx.QueryContext(ctx, `SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = $1`, table)
	if err != nil {
		slog.Default().Error(fmt.Sprintf("indexes %s: %v", table, err))
		return merr.Server("request failed")
	}
	var names []string
	for rows.Next() {
		var index string
		if err := rows.Scan(&index); err != nil {
			_ = rows.Close()
			return merr.Server("request failed")
		}
		names = append(names, index)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return merr.Server("request failed")
	}
	_ = rows.Close()
	for _, index := range names {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DROP INDEX IF EXISTS %q`, index)); err != nil {
			slog.Default().Error(fmt.Sprintf("drop index %s: %v", index, err))
			return merr.Server("request failed")
		}
	}
	return nil
}

func ensureWeekPartitions(ctx context.Context, db *sql.DB, name string, now time.Time) error {
	kind, err := relationKind(ctx, db, name)
	if err != nil {
		return err
	}
	if kind != "p" {
		return nil
	}
	start, _ := weekStart(now)
	for _, at := range []time.Time{start, start.AddDate(0, 0, 7)} {
		from, until := weekStart(at)
		part := fmt.Sprintf("%s_y%04dw%02d", name, from.Year(), isoWeek(from))
		statement := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			part, name, from.Format("2006-01-02"), until.Format("2006-01-02"),
		)
		if _, err := db.ExecContext(ctx, statement); err != nil {
			slog.Default().Error(fmt.Sprintf("ensure %s week: %v", name, err))
			return merr.Server("request failed")
		}
	}
	return nil
}

func ensureAlertEventWeeks(ctx context.Context, db *sql.DB, now time.Time) error {
	return ensureWeekPartitions(ctx, db, "alert_events", now)
}

func ensureMessageLogWeeks(ctx context.Context, db *sql.DB, now time.Time) error {
	return ensureWeekPartitions(ctx, db, "message_logs", now)
}

func relationKind(ctx context.Context, db *sql.DB, name string) (string, error) {
	var kind sql.NullString
	err := db.QueryRowContext(ctx, `SELECT c.relkind::text FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relname = $1`, name).Scan(&kind)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		slog.Default().Error(fmt.Sprintf("schema: %v", err))
		return "", merr.Server("request failed")
	}
	return kind.String, nil
}

func weekStart(at time.Time) (time.Time, time.Time) {
	at = at.UTC()
	weekday := int(at.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
	return start, start.AddDate(0, 0, 7)
}

func isoWeek(start time.Time) int {
	_, week := start.ISOWeek()
	return week
}
