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
	statements := []string{
		fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, name, heap),
		fmt.Sprintf(`CREATE TABLE %s (LIKE %s INCLUDING DEFAULTS INCLUDING IDENTITY) PARTITION BY RANGE (created_at)`, name, heap),
		fmt.Sprintf(`ALTER TABLE %s ADD PRIMARY KEY (created_at, id)`, name),
		fmt.Sprintf(`CREATE UNIQUE INDEX %s_uid_created_at ON %s (uid, created_at)`, name, name),
	}
	for _, statement := range statements {
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
