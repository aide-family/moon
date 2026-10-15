package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	entschema "entgo.io/ent/dialect/sql/schema"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/pkg/merr"
)

// ApplySchema 创建表，并把空的 alert_events 转成按 created_at 的周分区表。
// 已经分区时跳过这张表，避免生成器把主键改回单列 id。
func ApplySchema(ctx context.Context, client *ent.Client, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_events")
	if err != nil {
		return err
	}
	opts := []entschema.MigrateOption{}
	if kind == "p" {
		opts = append(opts, entschema.WithHooks(skipTable("alert_events")))
	}
	if err := client.Schema.Create(ctx, opts...); err != nil {
		return err
	}
	if err := partitionAlertEvents(ctx, db); err != nil {
		return err
	}
	return ensureAlertEventWeeks(ctx, db, time.Now().UTC())
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

func partitionAlertEvents(ctx context.Context, db *sql.DB) error {
	kind, err := relationKind(ctx, db, "alert_events")
	if err != nil {
		return err
	}
	if kind == "p" || kind == "" {
		return nil
	}
	if kind != "r" {
		return merr.Server("alert_events cannot be partitioned")
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM alert_events`).Scan(&count); err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("count alert_events: %v", err)
		return merr.Server("request failed")
	}
	if count > 0 {
		return merr.Server("alert_events is not empty")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return merr.Server("request failed")
	}
	defer func() { _ = tx.Rollback() }()
	statements := []string{
		`ALTER TABLE alert_events RENAME TO alert_events_heap`,
		`CREATE TABLE alert_events (LIKE alert_events_heap INCLUDING DEFAULTS INCLUDING IDENTITY) PARTITION BY RANGE (created_at)`,
		`ALTER TABLE alert_events ADD PRIMARY KEY (created_at, id)`,
		`CREATE UNIQUE INDEX alert_events_uid_created_at ON alert_events (uid, created_at)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			log.NewHelper(log.DefaultLogger).Errorf("partition alert_events: %v", err)
			return merr.Server("request failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE alert_events_heap`); err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("drop alert_events heap: %v", err)
		return merr.Server("request failed")
	}
	return tx.Commit()
}

func ensureAlertEventWeeks(ctx context.Context, db *sql.DB, now time.Time) error {
	kind, err := relationKind(ctx, db, "alert_events")
	if err != nil {
		return err
	}
	if kind != "p" {
		return nil
	}
	start, _ := weekStart(now)
	for _, at := range []time.Time{start, start.AddDate(0, 0, 7)} {
		from, until := weekStart(at)
		name := fmt.Sprintf("alert_events_y%04dw%02d", from.Year(), isoWeek(from))
		statement := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF alert_events FOR VALUES FROM ('%s') TO ('%s')`,
			name, from.Format("2006-01-02"), until.Format("2006-01-02"),
		)
		if _, err := db.ExecContext(ctx, statement); err != nil {
			log.NewHelper(log.DefaultLogger).Errorf("ensure alert_events week: %v", err)
			return merr.Server("request failed")
		}
	}
	return nil
}

func relationKind(ctx context.Context, db *sql.DB, name string) (string, error) {
	var kind sql.NullString
	err := db.QueryRowContext(ctx, `SELECT c.relkind::text FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relname = $1`, name).Scan(&kind)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("schema: %v", err)
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
