package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

// defaultMessageLogTablePrefix 与 config.yaml 的默认值一致。
const defaultMessageLogTablePrefix = "message_logs_"

// messageLogStore 是消息日志的分库分表存储。
// 与历史告警使用同一套分片原则：每个租户一个数据库（前缀 + 租户标识），
// 库内按 created_at 所在周建表（前缀 + 周一 YYYYMMDD，UTC）。
// 主库 message_logs 只在事务内承接写入，提交后归档到租户库并删除；
// 归档失败时保留主库行作为兜底，由调度器重试。
type messageLogStore struct {
	*shardStore
	tablePrefix string
}

func newMessageLogStore(shards *shardStore, tablePrefix string) *messageLogStore {
	logger := slog.Default()
	if shards != nil {
		logger = shards.log
	}
	if !validShardPrefix(tablePrefix, maxShardTablePrefixLen) {
		if tablePrefix != "" {
			logger.Error(fmt.Sprintf("invalid message log table prefix %q, using default", tablePrefix))
		}
		tablePrefix = defaultMessageLogTablePrefix
	}
	return &messageLogStore{shardStore: shards, tablePrefix: tablePrefix}
}

// tableName 返回某周的消息日志表名。week 会先归一到周一零点（UTC）。
func (s *messageLogStore) tableName(at time.Time) (string, time.Time) {
	return shardTableName(s.tablePrefix, at)
}

// ensureTable 在租户库里创建某一周的消息日志表。DDL 幂等，归档与回填都会调用。
func (s *messageLogStore) ensureTable(ctx context.Context, conn *sql.DB, table string) error {
	quoted := quoteShardIdent(table)
	statements := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			uid VARCHAR(64) NOT NULL PRIMARY KEY,
			namespace_uid VARCHAR(64) NOT NULL,
			task_uid VARCHAR(64) NOT NULL,
			status VARCHAR(32) NOT NULL,
			channel_uid VARCHAR(64) NOT NULL DEFAULT '',
			template_uid VARCHAR(64) NOT NULL DEFAULT '',
			sender TEXT NOT NULL DEFAULT '',
			receivers TEXT NOT NULL DEFAULT '',
			subject TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			detail TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		)`, quoted),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (namespace_uid, created_at)`,
			quoteShardIdent(table+"_ns_created"), quoted),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (namespace_uid, status, created_at)`,
			quoteShardIdent(table+"_ns_status_created"), quoted),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (task_uid)`,
			quoteShardIdent(table+"_task_uid"), quoted),
	}
	for _, statement := range statements {
		if _, err := conn.ExecContext(ctx, statement); err != nil {
			s.log.Error(fmt.Sprintf("message log table %s: %v", table, err))
			return merr.Server("message log storage is unavailable")
		}
	}
	return nil
}

// archive 把一条消息日志写入租户库的周表。重复归档按 uid 覆盖。
func (s *messageLogStore) archive(ctx context.Context, item bo.MessageLog) (string, string, time.Time, error) {
	at := item.CreatedAt
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
		uid, namespace_uid, task_uid, status, channel_uid, template_uid,
		sender, receivers, subject, content, detail, created_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	ON CONFLICT (uid) DO UPDATE SET
		namespace_uid = EXCLUDED.namespace_uid,
		task_uid = EXCLUDED.task_uid,
		status = EXCLUDED.status,
		channel_uid = EXCLUDED.channel_uid,
		template_uid = EXCLUDED.template_uid,
		sender = EXCLUDED.sender,
		receivers = EXCLUDED.receivers,
		subject = EXCLUDED.subject,
		content = EXCLUDED.content,
		detail = EXCLUDED.detail,
		created_at = EXCLUDED.created_at`, quoteShardIdent(table)),
		item.UID, item.NamespaceUID, item.TaskUID, item.Status, item.ChannelUID, item.TemplateUID,
		item.Sender, item.Receivers, item.Subject, item.Content, item.Detail, at,
	)
	if err != nil {
		s.log.Error(fmt.Sprintf("archive message log %s: %v", item.UID, err))
		return "", "", time.Time{}, merr.Server("message log storage is unavailable")
	}
	return database, table, week, nil
}

// list 读取租户库中命中筛选条件的周表。
// limit 限制返回行数（按 created_at 倒序），total 是命中条件的总行数。
// 没有租户库时返回空结果，不创建空库。
func (s *messageLogStore) list(ctx context.Context, namespaceUID string, filter bo.MessageLogListFilter, limit int) ([]bo.MessageLog, int, error) {
	if s == nil || s.shardStore == nil || s.admin == nil {
		return nil, 0, nil
	}
	exists, err := s.tenantExists(ctx, namespaceUID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, nil
	}
	conn, _, err := s.ensureTenant(ctx, namespaceUID)
	if err != nil {
		return nil, 0, err
	}
	tables, err := s.tables(ctx, conn, filter.StartAt, filter.EndAt)
	if err != nil {
		return nil, 0, err
	}
	if len(tables) == 0 {
		return nil, 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	where, args := messageLogShardWhere(namespaceUID, filter)
	branches := make([]string, 0, len(tables))
	for _, table := range tables {
		branches = append(branches, fmt.Sprintf(`SELECT %s FROM %s WHERE %s`,
			messageLogShardColumns, quoteShardIdent(table), where))
	}
	union := strings.Join(branches, " UNION ALL ")
	var total int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM (`+union+`) AS shard_logs`, args...).Scan(&total); err != nil {
		s.log.Error(fmt.Sprintf("count message logs %s: %v", namespaceUID, err))
		return nil, 0, merr.Server("message log storage is unavailable")
	}
	if total == 0 {
		return nil, 0, nil
	}
	args = append(args, limit)
	query := union + fmt.Sprintf(` ORDER BY created_at DESC, uid DESC LIMIT $%d`, len(args))
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Error(fmt.Sprintf("query message logs %s: %v", namespaceUID, err))
		return nil, 0, merr.Server("message log storage is unavailable")
	}
	defer rows.Close()
	items := make([]bo.MessageLog, 0)
	for rows.Next() {
		item, err := scanMessageLog(rows)
		if err != nil {
			s.log.Error(fmt.Sprintf("scan message logs %s: %v", namespaceUID, err))
			return nil, 0, merr.Server("message log storage is unavailable")
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		s.log.Error(fmt.Sprintf("message log rows %s: %v", namespaceUID, err))
		return nil, 0, merr.Server("message log storage is unavailable")
	}
	return items, total, nil
}

// get 按分片索引读取一条消息日志。database/table 由主库索引提供。
func (s *messageLogStore) get(ctx context.Context, namespaceUID, database, table, uid string) (*bo.MessageLog, error) {
	if !shardIdentPattern.MatchString(database) || !shardIdentPattern.MatchString(table) {
		return nil, merr.NotFound("not found")
	}
	conn, err := s.connByName(database)
	if err != nil {
		return nil, err
	}
	row := conn.QueryRowContext(ctx, fmt.Sprintf(`SELECT %s FROM %s WHERE namespace_uid = $1 AND uid = $2`,
		messageLogShardColumns, quoteShardIdent(table)), namespaceUID, uid)
	item, err := scanMessageLog(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, merr.NotFound("not found")
		}
		s.log.Error(fmt.Sprintf("message log get %s: %v", table, err))
		return nil, merr.Server("message log storage is unavailable")
	}
	return &item, nil
}

// tables 列出与时间范围相交的消息日志周表。
// 周表覆盖 [周一, 下周一)，所以只要与 [start, end) 有交集就需要读取。
func (s *messageLogStore) tables(ctx context.Context, conn *sql.DB, start, end time.Time) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE $1 ORDER BY tablename`, s.tablePrefix+"%")
	if err != nil {
		s.log.Error(fmt.Sprintf("message log tables: %v", err))
		return nil, merr.Server("message log storage is unavailable")
	}
	names := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			_ = rows.Close()
			return nil, merr.Server("message log storage is unavailable")
		}
		names = append(names, table)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, merr.Server("message log storage is unavailable")
	}
	_ = rows.Close()
	return filterMessageLogTables(names, s.tablePrefix, start, end), nil
}

// filterMessageLogTables 只保留与时间范围相交、且名字合法的周表。
func filterMessageLogTables(names []string, tablePrefix string, start, end time.Time) []string {
	tables := make([]string, 0, len(names))
	for _, name := range names {
		if !shardIdentPattern.MatchString(name) || !strings.HasPrefix(name, tablePrefix) {
			continue
		}
		week, err := time.ParseInLocation("20060102", strings.TrimPrefix(name, tablePrefix), time.UTC)
		if err != nil {
			continue
		}
		if !start.IsZero() && !week.AddDate(0, 0, 7).After(start) {
			continue
		}
		if !end.IsZero() && !week.Before(end) {
			continue
		}
		tables = append(tables, name)
	}
	return tables
}

const messageLogShardColumns = `uid, namespace_uid, task_uid, status, channel_uid, template_uid,
	sender, receivers, subject, content, detail, created_at`

// messageLogShardWhere 生成周表共用的筛选条件；参数按顺序复用到每个 UNION 分支。
func messageLogShardWhere(namespaceUID string, filter bo.MessageLogListFilter) (string, []any) {
	args := []any{namespaceUID}
	where := "namespace_uid = $1"
	if filter.TaskUID != "" {
		args = append(args, filter.TaskUID)
		where += fmt.Sprintf(" AND task_uid = $%d", len(args))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if !filter.StartAt.IsZero() {
		args = append(args, filter.StartAt)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if !filter.EndAt.IsZero() {
		args = append(args, filter.EndAt)
		where += fmt.Sprintf(" AND created_at < $%d", len(args))
	}
	return where, args
}

func scanMessageLog(row historyScanner) (bo.MessageLog, error) {
	var item bo.MessageLog
	err := row.Scan(
		&item.UID, &item.NamespaceUID, &item.TaskUID, &item.Status, &item.ChannelUID,
		&item.TemplateUID, &item.Sender, &item.Receivers, &item.Subject, &item.Content,
		&item.Detail, &item.CreatedAt,
	)
	if err != nil {
		return bo.MessageLog{}, err
	}
	return item, nil
}

// mergeMessageLogs 合并主库兜底行与租户周表行，按 UID 去重并按创建时间倒序。
// 周表数据优先，主库行只在归档未完成时存在。
func mergeMessageLogs(sharded, fallback []bo.MessageLog, limit int) []bo.MessageLog {
	seen := make(map[string]struct{}, len(sharded)+len(fallback))
	items := make([]bo.MessageLog, 0, len(sharded)+len(fallback))
	for _, group := range [][]bo.MessageLog{sharded, fallback} {
		for _, item := range group {
			if _, ok := seen[item.UID]; ok {
				continue
			}
			seen[item.UID] = struct{}{}
			items = append(items, item)
		}
	}
	sortMessageLogs(items)
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func sortMessageLogs(items []bo.MessageLog) {
	sort.Slice(items, func(i, j int) bool {
		return messageLogBefore(items[i], items[j])
	})
}

// messageLogBefore 定义列表的稳定顺序：创建时间新的在前，同一秒内按 UID 倒序。
func messageLogBefore(left, right bo.MessageLog) bool {
	if !left.CreatedAt.Equal(right.CreatedAt) {
		return left.CreatedAt.After(right.CreatedAt)
	}
	return left.UID > right.UID
}
