package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// defaultHistoryDatabasePrefix 与 config.yaml 的默认值一致。
	defaultHistoryDatabasePrefix = "alioth_alert_"
	// defaultHistoryTablePrefix 与 config.yaml 的默认值一致。
	defaultHistoryTablePrefix = "alert_events_"
	// historyIdentMaxLen 是 PostgreSQL 标识符的长度上限。
	historyIdentMaxLen = 63
)

var historyIdentPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// historyStore 是历史告警的分库分表存储。
// 每个租户一个数据库（前缀 + 租户标识），库内按周一张表（前缀 + 周一 YYYYMMDD）。
// 主库只保留 alert_history_events 路由索引和仍在触发的事件。
type historyStore struct {
	baseSource  string
	admin       *sql.DB
	dbPrefix    string
	tablePrefix string
	log         *slog.Logger

	mu    sync.Mutex
	conns map[string]*sql.DB
}

func newHistoryStore(admin *sql.DB, source, dbPrefix, tablePrefix string, logger *slog.Logger) *historyStore {
	if logger == nil {
		logger = slog.Default()
	}
	if !validHistoryPrefix(dbPrefix) {
		if dbPrefix != "" {
			logger.Error(fmt.Sprintf("invalid history database prefix %q, using default", dbPrefix))
		}
		dbPrefix = defaultHistoryDatabasePrefix
	}
	if !validHistoryPrefix(tablePrefix) {
		if tablePrefix != "" {
			logger.Error(fmt.Sprintf("invalid history table prefix %q, using default", tablePrefix))
		}
		tablePrefix = defaultHistoryTablePrefix
	}
	return &historyStore{
		baseSource: source, admin: admin, dbPrefix: dbPrefix, tablePrefix: tablePrefix,
		log: logger, conns: make(map[string]*sql.DB),
	}
}

func validHistoryPrefix(value string) bool {
	return value != "" && len(value) < historyIdentMaxLen && historyIdentPattern.MatchString(value)
}

func (s *historyStore) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, conn := range s.conns {
		if err := conn.Close(); err != nil {
			s.log.Error(fmt.Sprintf("close history database %s: %v", name, err))
		}
	}
	s.conns = make(map[string]*sql.DB)
}

// databaseName 返回租户库名。租户标识允许被截断或清洗，保证结果仍是合法标识符且不超过 63 字节。
func (s *historyStore) databaseName(namespaceUID string) string {
	return historyDatabaseName(s.dbPrefix, namespaceUID)
}

func historyDatabaseName(prefix, namespaceUID string) string {
	tenant := sanitizeHistoryIdent(namespaceUID)
	if tenant == "" {
		tenant = "default"
	}
	name := prefix + tenant
	if len(name) <= historyIdentMaxLen {
		return name
	}
	// 超长时保留前缀和摘要，避免不同租户截断后撞名。
	sum := sha256.Sum256([]byte(namespaceUID))
	suffix := hex.EncodeToString(sum[:4])
	keep := historyIdentMaxLen - len(prefix) - len(suffix) - 1
	if keep < 1 {
		keep = 1
	}
	if len(tenant) > keep {
		tenant = tenant[:keep]
	}
	return prefix + tenant + "_" + suffix
}

// tableName 返回某周的表名。week 会先归一到周一零点（UTC）。
func (s *historyStore) tableName(at time.Time) (string, time.Time) {
	week := historyWeekStart(at)
	return s.tablePrefix + week.Format("20060102"), week
}

func historyWeekStart(at time.Time) time.Time {
	at = at.UTC()
	weekday := int(at.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
}

func sanitizeHistoryIdent(value string) string {
	var out strings.Builder
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			out.WriteRune(r)
		default:
			// 租户标识通常是十六进制；其他字符统一替换，避免拼出非法或带引号的库名。
			out.WriteByte('_')
		}
	}
	return out.String()
}

func quoteHistoryIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

// ensureTenant 确保租户库存在并返回连接。CREATE DATABASE 只在主库执行一次；连接按租户缓存。
func (s *historyStore) ensureTenant(ctx context.Context, namespaceUID string) (*sql.DB, string, error) {
	name := s.databaseName(namespaceUID)
	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		s.mu.Unlock()
		return conn, name, nil
	}
	s.mu.Unlock()

	var exists bool
	if err := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&exists); err != nil {
		s.log.Error(fmt.Sprintf("history database check %s: %v", name, err))
		return nil, name, merr.Server("history storage is unavailable")
	}
	if !exists {
		if _, err := s.admin.ExecContext(ctx, `CREATE DATABASE `+quoteHistoryIdent(name)); err != nil {
			// 并发创建时另一个副本可能已经建好，重新确认一次再决定是否失败。
			var nowExists bool
			if checkErr := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&nowExists); checkErr != nil || !nowExists {
				s.log.Error(fmt.Sprintf("create history database %s: %v", name, err))
				return nil, name, merr.Server("history storage is unavailable")
			}
		}
	}

	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		s.mu.Unlock()
		return conn, name, nil
	}
	s.mu.Unlock()

	conn, err := s.connByName(name)
	if err != nil {
		return nil, name, err
	}
	return conn, name, nil
}

// connByName 打开并缓存指定历史库连接。用于按路由索引读取旧前缀的库。
func (s *historyStore) connByName(name string) (*sql.DB, error) {
	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		s.mu.Unlock()
		return conn, nil
	}
	s.mu.Unlock()
	source, err := historyDSN(s.baseSource, name)
	if err != nil {
		s.log.Error(fmt.Sprintf("history database dsn %s: %v", name, err))
		return nil, merr.Server("history storage is unavailable")
	}
	conn, err := openHistoryDB(source, s.log)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	// 其他请求可能已经先建好连接，丢掉本次多余的一条。
	if existing, ok := s.conns[name]; ok {
		s.mu.Unlock()
		_ = conn.Close()
		return existing, nil
	}
	s.conns[name] = conn
	s.mu.Unlock()
	return conn, nil
}

// ensureTable 在租户库里创建某一周的周表。DDL 是幂等的，只由恢复写路径和归档调用。
func (s *historyStore) ensureTable(ctx context.Context, conn *sql.DB, table string) error {
	quoted := quoteHistoryIdent(table)
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
			quoteHistoryIdent(table+"_ns_created"), quoted),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (status)`,
			quoteHistoryIdent(table+"_status"), quoted),
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
		created_at = EXCLUDED.created_at`, quoteHistoryIdent(table)),
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
		if !historyIdentPattern.MatchString(table) {
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
		FROM %s WHERE namespace_uid = $1 ORDER BY created_at DESC`, quoteHistoryIdent(table)), namespaceUID)
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
	if !historyIdentPattern.MatchString(table) || !historyIdentPattern.MatchString(database) {
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
		FROM %s WHERE namespace_uid = $1 AND uid = $2`, quoteHistoryIdent(table)), namespaceUID, uid)
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
	name := s.databaseName(namespaceUID)
	var exists bool
	if err := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&exists); err != nil {
		return 0, merr.Server("history storage is unavailable")
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
		if !historyIdentPattern.MatchString(table) {
			continue
		}
		var count int
		if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM `+quoteHistoryIdent(table)+` WHERE namespace_uid = $1`, namespaceUID).Scan(&count); err != nil {
			return 0, merr.Server("history storage is unavailable")
		}
		total += count
	}
	return total, nil
}

// dropTenant 只在测试清理时使用，删除租户的历史库。
func (s *historyStore) dropTenant(ctx context.Context, namespaceUID string) error {
	name := s.databaseName(namespaceUID)
	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		_ = conn.Close()
		delete(s.conns, name)
	}
	s.mu.Unlock()
	_, err := s.admin.ExecContext(ctx, `DROP DATABASE IF EXISTS `+quoteHistoryIdent(name))
	return err
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

// historyDSN 把主库连接串里的库名换成租户历史库名。支持 URL 和 key=value 两种写法。
func historyDSN(source, database string) (string, error) {
	if parsed, err := url.Parse(source); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		copied := *parsed
		copied.Path = "/" + database
		return copied.String(), nil
	}
	fields := strings.Fields(source)
	out := make([]string, 0, len(fields)+1)
	replaced := false
	for _, field := range fields {
		key, _, ok := strings.Cut(field, "=")
		if ok && (key == "dbname" || key == "database") {
			out = append(out, key+"="+database)
			replaced = true
			continue
		}
		out = append(out, field)
	}
	if !replaced {
		out = append(out, "dbname="+database)
	}
	return strings.Join(out, " "), nil
}

// openHistoryDB 为历史库使用更小的连接池。一个租户一个库，连接数按租户数放大，
// 不能沿用主库单副本 20 个连接的上限。
func openHistoryDB(source string, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", source)
	if err != nil {
		return nil, databaseErr(logger, err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, databaseErr(logger, err)
	}
	return db, nil
}
