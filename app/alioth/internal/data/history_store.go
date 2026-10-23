package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

// defaultHistoryTablePrefix 与 config.yaml 的默认值一致。
const defaultHistoryTablePrefix = "alert_events_"

// historyStore 是历史告警的分库分表存储。
// 每个租户一个数据库（前缀 + 租户标识），库内按周一张表（前缀 + 周一 YYYYMMDD）。
// 租户库和连接由 shardStore 统一管理，消息日志共用同一套租户库。
// 主库只保留 alert_history_events 路由索引和仍在触发的事件。
type historyStore struct {
	*shardStore
	tablePrefix string
}

func newHistoryStore(shards *shardStore, tablePrefix string) *historyStore {
	logger := slog.Default()
	if shards != nil {
		logger = shards.log
	}
	if !validShardPrefix(tablePrefix, maxShardTablePrefixLen) {
		if tablePrefix != "" {
			logger.Error(fmt.Sprintf("invalid history table prefix %q, using default", tablePrefix))
		}
		tablePrefix = defaultHistoryTablePrefix
	}
	return &historyStore{shardStore: shards, tablePrefix: tablePrefix}
}

// tableName 返回某周的表名。week 会先归一到周一零点（UTC）。
func (s *historyStore) tableName(at time.Time) (string, time.Time) {
	return shardTableName(s.tablePrefix, at)
}

// ensureTable 在租户库里创建某一周的周表。DDL 是幂等的，只由恢复写路径和归档调用。
func (s *historyStore) ensureTable(ctx context.Context, conn *sql.DB, table string) error {
	quoted := quoteShardIdent(table)
	statements := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			uid VARCHAR(64) NOT NULL PRIMARY KEY,
			namespace_uid VARCHAR(64) NOT NULL,
			rule_uid VARCHAR(64) NOT NULL DEFAULT '',
			datasource_uid VARCHAR(64) NOT NULL DEFAULT '',
			fingerprint VARCHAR(128) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL,
			expr TEXT NOT NULL DEFAULT '',
			labels TEXT NOT NULL DEFAULT '{}',
			summary TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			sample_value TEXT NOT NULL DEFAULT '',
			level_uid VARCHAR(64) NOT NULL DEFAULT '',
			starts_at TIMESTAMPTZ NOT NULL,
			ends_at TIMESTAMPTZ NULL,
			eval_slot TIMESTAMPTZ NOT NULL,
			suppressed BOOLEAN NOT NULL DEFAULT FALSE,
			suppressed_until TIMESTAMPTZ NULL,
			intervened_by VARCHAR(64) NOT NULL DEFAULT '',
			intervened_at TIMESTAMPTZ NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`, quoted),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (namespace_uid, created_at)`,
			quoteShardIdent(table+"_ns_created"), quoted),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (status)`,
			quoteShardIdent(table+"_status"), quoted),
	}
	for _, statement := range statements {
		if _, err := conn.ExecContext(ctx, statement); err != nil {
			s.log.Error(fmt.Sprintf("history table %s: %v", table, err))
			return merr.Server("history storage is unavailable")
		}
	}
	return nil
}

// archive 把一条已恢复事件写入租户库的周表。重复归档按 uid 覆盖。
func (s *historyStore) archive(ctx context.Context, item bo.AlertEvent) (string, string, time.Time, error) {
	at := item.CreatedAt
	if at.IsZero() {
		at = item.StartsAt
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	table, week := s.tableName(at)
	conn, database, err := s.ensureTenant(ctx, item.NamespaceUID)
	if err != nil {
		return "", "", time.Time{}, err
	}
	if err := s.ensureTable(ctx, conn, table); err != nil {
		return "", "", time.Time{}, err
	}
	_, err = conn.ExecContext(ctx, fmt.Sprintf(`INSERT INTO %s (
		uid, namespace_uid, rule_uid, datasource_uid, fingerprint, status, expr, labels,
		summary, description, sample_value, level_uid, starts_at, ends_at, eval_slot,
		suppressed, suppressed_until, intervened_by, intervened_at, created_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
	ON CONFLICT (uid) DO UPDATE SET
		namespace_uid = EXCLUDED.namespace_uid,
		rule_uid = EXCLUDED.rule_uid,
		datasource_uid = EXCLUDED.datasource_uid,
		fingerprint = EXCLUDED.fingerprint,
		status = EXCLUDED.status,
		expr = EXCLUDED.expr,
		labels = EXCLUDED.labels,
		summary = EXCLUDED.summary,
		description = EXCLUDED.description,
		sample_value = EXCLUDED.sample_value,
		level_uid = EXCLUDED.level_uid,
		starts_at = EXCLUDED.starts_at,
		ends_at = EXCLUDED.ends_at,
		eval_slot = EXCLUDED.eval_slot,
		suppressed = EXCLUDED.suppressed,
		suppressed_until = EXCLUDED.suppressed_until,
		intervened_by = EXCLUDED.intervened_by,
		intervened_at = EXCLUDED.intervened_at,
		created_at = EXCLUDED.created_at`, quoteShardIdent(table)),
		item.UID, item.NamespaceUID, item.RuleUID, item.DatasourceUID, item.Fingerprint, item.Status,
		item.Expr, item.Labels, item.Summary, item.Description, item.SampleValue, item.LevelUID,
		item.StartsAt, item.EndsAt, item.EvalSlot, item.Suppressed, item.SuppressedUntil,
		item.IntervenedBy, item.IntervenedAt, item.CreatedAt,
	)
	if err != nil {
		s.log.Error(fmt.Sprintf("archive history event %s: %v", item.UID, err))
		return "", "", time.Time{}, merr.Server("history storage is unavailable")
	}
	return database, table, week, nil
}

// list 读取租户所有历史周表。没有租户库时返回空列表，不创建空库。
func (s *historyStore) list(ctx context.Context, namespaceUID string) ([]bo.AlertEvent, error) {
	name := s.databaseName(namespaceUID)
	var exists bool
	if err := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&exists); err != nil {
		s.log.Error(fmt.Sprintf("history database check %s: %v", name, err))
		return nil, merr.Server("history storage is unavailable")
	}
	if !exists {
		return nil, nil
	}
	conn, _, err := s.ensureTenant(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	rows, err := conn.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE $1 ORDER BY tablename`, s.tablePrefix+"%")
	if err != nil {
		s.log.Error(fmt.Sprintf("history tables %s: %v", name, err))
		return nil, merr.Server("history storage is unavailable")
	}
	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			_ = rows.Close()
			return nil, merr.Server("history storage is unavailable")
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, merr.Server("history storage is unavailable")
	}
	_ = rows.Close()

	items := make([]bo.AlertEvent, 0)
	for _, table := range tables {
		if !shardIdentPattern.MatchString(table) {
			continue
		}
		part, err := s.queryTable(ctx, conn, table, namespaceUID)
		if err != nil {
			return nil, err
		}
		items = append(items, part...)
	}
	return items, nil
}

func (s *historyStore) queryTable(ctx context.Context, conn *sql.DB, table, namespaceUID string) ([]bo.AlertEvent, error) {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf(`SELECT
		uid, namespace_uid, rule_uid, datasource_uid, fingerprint, status, expr, labels,
		summary, description, sample_value, level_uid, starts_at, ends_at, eval_slot,
		suppressed, suppressed_until, intervened_by, intervened_at, created_at
		FROM %s WHERE namespace_uid = $1 ORDER BY created_at DESC`, quoteShardIdent(table)), namespaceUID)
	if err != nil {
		s.log.Error(fmt.Sprintf("history query %s: %v", table, err))
		return nil, merr.Server("history storage is unavailable")
	}
	defer rows.Close()
	items := make([]bo.AlertEvent, 0)
	for rows.Next() {
		item, err := scanHistoryEvent(rows)
		if err != nil {
			s.log.Error(fmt.Sprintf("history scan %s: %v", table, err))
			return nil, merr.Server("history storage is unavailable")
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		s.log.Error(fmt.Sprintf("history rows %s: %v", table, err))
		return nil, merr.Server("history storage is unavailable")
	}
	return items, nil
}

// get 按索引导航读取一条历史事件。table 为空表示索引缺失，返回找不到。
func (s *historyStore) get(ctx context.Context, namespaceUID, database, table, uid string) (*bo.AlertEvent, error) {
	if !shardIdentPattern.MatchString(table) || !shardIdentPattern.MatchString(database) {
		return nil, merr.NotFound("not found")
	}
	conn, err := s.connByName(database)
	if err != nil {
		return nil, err
	}
	row := conn.QueryRowContext(ctx, fmt.Sprintf(`SELECT
		uid, namespace_uid, rule_uid, datasource_uid, fingerprint, status, expr, labels,
		summary, description, sample_value, level_uid, starts_at, ends_at, eval_slot,
		suppressed, suppressed_until, intervened_by, intervened_at, created_at
		FROM %s WHERE namespace_uid = $1 AND uid = $2`, quoteShardIdent(table)), namespaceUID, uid)
	item, err := scanHistoryEvent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, merr.NotFound("not found")
		}
		s.log.Error(fmt.Sprintf("history get %s: %v", table, err))
		return nil, merr.Server("history storage is unavailable")
	}
	return &item, nil
}

// count 只统计行数，用于历史统计条，不读取事件正文。
func (s *historyStore) count(ctx context.Context, namespaceUID string) (int, error) {
	exists, err := s.tenantExists(ctx, namespaceUID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}
	conn, _, err := s.ensureTenant(ctx, namespaceUID)
	if err != nil {
		return 0, err
	}
	rows, err := conn.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE $1`, s.tablePrefix+"%")
	if err != nil {
		return 0, merr.Server("history storage is unavailable")
	}
	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			_ = rows.Close()
			return 0, merr.Server("history storage is unavailable")
		}
		tables = append(tables, table)
	}
	_ = rows.Close()
	total := 0
	for _, table := range tables {
		if !shardIdentPattern.MatchString(table) {
			continue
		}
		var count int
		if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM `+quoteShardIdent(table)+` WHERE namespace_uid = $1`, namespaceUID).Scan(&count); err != nil {
			return 0, merr.Server("history storage is unavailable")
		}
		total += count
	}
	return total, nil
}

type historyScanner interface {
	Scan(dest ...any) error
}

func scanHistoryEvent(row historyScanner) (bo.AlertEvent, error) {
	var item bo.AlertEvent
	var endsAt, suppressedUntil, intervenedAt sql.NullTime
	err := row.Scan(
		&item.UID, &item.NamespaceUID, &item.RuleUID, &item.DatasourceUID, &item.Fingerprint,
		&item.Status, &item.Expr, &item.Labels, &item.Summary, &item.Description,
		&item.SampleValue, &item.LevelUID, &item.StartsAt, &endsAt, &item.EvalSlot,
		&item.Suppressed, &suppressedUntil, &item.IntervenedBy, &intervenedAt, &item.CreatedAt,
	)
	if err != nil {
		return bo.AlertEvent{}, err
	}
	if endsAt.Valid {
		item.EndsAt = &endsAt.Time
	}
	if suppressedUntil.Valid {
		item.SuppressedUntil = &suppressedUntil.Time
	}
	if intervenedAt.Valid {
		item.IntervenedAt = &intervenedAt.Time
	}
	return item, nil
}
