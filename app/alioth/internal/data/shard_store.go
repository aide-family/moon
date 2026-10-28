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

	"github.com/aide-family/moon/pkg/merr"
)

const (
	// defaultShardDatabasePrefix 与 config.yaml 的默认值一致。
	defaultShardDatabasePrefix = "alioth_alert_"
	// shardIdentMaxLen 是 PostgreSQL 标识符的长度上限。
	shardIdentMaxLen = 63
	// id.New 生成 16 字节随机数的十六进制 UID。
	shardNamespaceUIDLen = 32
	// YYYYMMDD 周日期后缀的长度。
	shardWeekDateLen = 8
	// 前缀长度确保拼接完整租户 UID 或周日期后不超过 PostgreSQL 上限。
	maxShardDatabasePrefixLen = shardIdentMaxLen - shardNamespaceUIDLen
	maxShardTablePrefixLen    = shardIdentMaxLen - shardWeekDateLen
)

var shardIdentPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// shardStore 是分库分表的租户库基础设施：每个租户一个数据库，库内按周一张表。
// 告警历史和消息日志共用同一套租户库与命名规则（库前缀 + 租户标识，表前缀 + 周一日期），
// 因此共用连接缓存，避免同一租户库被打开多份连接池。
type shardStore struct {
	baseSource string
	admin      *sql.DB
	dbPrefix   string
	log        *slog.Logger

	mu    sync.Mutex
	conns map[string]*sql.DB
}

func newShardStore(admin *sql.DB, source, dbPrefix string, logger *slog.Logger) *shardStore {
	if logger == nil {
		logger = slog.Default()
	}
	if !validShardPrefix(dbPrefix, maxShardDatabasePrefixLen) {
		if dbPrefix != "" {
			logger.Error(fmt.Sprintf("invalid shard database prefix %q, using default", dbPrefix))
		}
		dbPrefix = defaultShardDatabasePrefix
	}
	return &shardStore{
		baseSource: source, admin: admin, dbPrefix: dbPrefix,
		log: logger, conns: make(map[string]*sql.DB),
	}
}

func validShardPrefix(value string, maxLen int) bool {
	return value != "" && len(value) <= maxLen && shardIdentPattern.MatchString(value)
}

func (s *shardStore) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, conn := range s.conns {
		if err := conn.Close(); err != nil {
			s.log.Error(fmt.Sprintf("close shard database %s: %v", name, err))
		}
	}
	s.conns = make(map[string]*sql.DB)
}

// databaseName 返回租户库名。租户标识允许被截断或清洗，保证结果仍是合法标识符且不超过 63 字节。
func (s *shardStore) databaseName(namespaceUID string) string {
	return shardDatabaseName(s.dbPrefix, namespaceUID)
}

func shardDatabaseName(prefix, namespaceUID string) string {
	tenant := sanitizeShardIdent(namespaceUID)
	if tenant == "" {
		tenant = "default"
	}
	name := prefix + tenant
	if len(name) <= shardIdentMaxLen {
		return name
	}
	// 超长时保留前缀和摘要，避免不同租户截断后撞名。
	sum := sha256.Sum256([]byte(namespaceUID))
	suffix := hex.EncodeToString(sum[:4])
	keep := shardIdentMaxLen - len(prefix) - len(suffix) - 1
	if keep < 1 {
		keep = 1
	}
	if len(tenant) > keep {
		tenant = tenant[:keep]
	}
	return prefix + tenant + "_" + suffix
}

// tableName 返回某周的表名。week 会先归一到周一零点（UTC）。
func shardTableName(prefix string, at time.Time) (string, time.Time) {
	week := shardWeekStart(at)
	return prefix + week.Format("20060102"), week
}

func shardWeekStart(at time.Time) time.Time {
	at = at.UTC()
	weekday := int(at.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
}

func sanitizeShardIdent(value string) string {
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

func quoteShardIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

// tenantExists 只检查租户库是否已经创建，避免只读请求顺带建空库。
func (s *shardStore) tenantExists(ctx context.Context, namespaceUID string) (bool, error) {
	name := s.databaseName(namespaceUID)
	var exists bool
	if err := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&exists); err != nil {
		s.log.Error(fmt.Sprintf("shard database check %s: %v", name, err))
		return false, merr.Server("shard storage is unavailable")
	}
	return exists, nil
}

// ensureTenant 确保租户库存在并返回连接。CREATE DATABASE 只在主库执行一次；连接按租户缓存。
func (s *shardStore) ensureTenant(ctx context.Context, namespaceUID string) (*sql.DB, string, error) {
	name := s.databaseName(namespaceUID)
	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		s.mu.Unlock()
		return conn, name, nil
	}
	s.mu.Unlock()

	var exists bool
	if err := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&exists); err != nil {
		s.log.Error(fmt.Sprintf("shard database check %s: %v", name, err))
		return nil, name, merr.Server("shard storage is unavailable")
	}
	if !exists {
		if _, err := s.admin.ExecContext(ctx, `CREATE DATABASE `+quoteShardIdent(name)); err != nil {
			// 并发创建时另一个副本可能已经建好，重新确认一次再决定是否失败。
			var nowExists bool
			if checkErr := s.admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, name).Scan(&nowExists); checkErr != nil || !nowExists {
				s.log.Error(fmt.Sprintf("create shard database %s: %v", name, err))
				return nil, name, merr.Server("shard storage is unavailable")
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

// connByName 打开并缓存指定租户库连接。用于按路由索引读取旧前缀的库。
func (s *shardStore) connByName(name string) (*sql.DB, error) {
	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		s.mu.Unlock()
		return conn, nil
	}
	s.mu.Unlock()
	source, err := shardDSN(s.baseSource, name)
	if err != nil {
		s.log.Error(fmt.Sprintf("shard database dsn %s: %v", name, err))
		return nil, merr.Server("shard storage is unavailable")
	}
	conn, err := openShardDB(source, s.log)
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

// dropTenant 只在测试清理时使用，删除租户的分片库。
func (s *shardStore) dropTenant(ctx context.Context, namespaceUID string) error {
	name := s.databaseName(namespaceUID)
	s.mu.Lock()
	if conn, ok := s.conns[name]; ok {
		_ = conn.Close()
		delete(s.conns, name)
	}
	s.mu.Unlock()
	_, err := s.admin.ExecContext(ctx, `DROP DATABASE IF EXISTS `+quoteShardIdent(name))
	return err
}

// shardDSN 把主库连接串里的库名换成租户分片库名。支持 URL 和 key=value 两种写法。
func shardDSN(source, database string) (string, error) {
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

// openShardDB 为分片库使用更小的连接池。一个租户一个库，连接数按租户数放大，
// 不能沿用主库单副本 20 个连接的上限。
func openShardDB(source string, logger *slog.Logger) (*sql.DB, error) {
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
