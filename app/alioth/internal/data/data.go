package data

import (
	_ "github.com/jackc/pgx/v5/stdlib"

	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(
	NewDependencies,
	NewUserRepo,
	NewCaptchaRepo,
	NewEmailCodeRepo,
	NewOAuthRepo,
	NewNamespaceRepo,
	NewMemberRepo,
	NewInviteRepo,
	NewMessageRepo,
	NewRoleRepo,
	NewDatasourceRepo,
	NewProbeRepo,
	NewAlertDictionaryRepo,
	NewAlertRuleRepo,
	NewAlertEventRepo,
	NewAlertPageRepo,
	NewNotificationRepo,
	NewWorkerRepo,
	NewAgentRepo,
	NewSSHCommandRepo,
	NewClusterRepo,
	NewKubernetesRepo,
	NewLockClient,
	NewReadinessChecker,
	NewPartitionRepo,
	NewFeatureRepo,
	NewNamespaceOrderRepo,
)

// Dependencies 持有数据层共享的数据库连接、ent 客户端、加密盒、分片存储
// 和 Kubernetes 客户端缓存。
type Dependencies struct {
	db          *sql.DB
	client      *ent.Client
	box         *secretBox
	history     *historyStore
	messageLogs *messageLogStore
	shards      *shardStore
	clients     *clusterClients
}

func NewDependencies(c *conf.Data, auth *conf.Auth, logger *slog.Logger) (*Dependencies, func(), error) {
	if c == nil || c.Database == nil || c.Database.Source == "" {
		return nil, nil, merr.Server("database source is empty")
	}
	if auth == nil || auth.GetEncryptKey() == "" {
		return nil, nil, merr.Server("encrypt key is empty")
	}
	box, err := newSecretBox(auth.GetEncryptKey())
	if err != nil {
		return nil, nil, err
	}
	helper := logger
	db, err := openDB(c.Database.Source, helper)
	if err != nil {
		return nil, nil, err
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	// 告警历史和消息日志共用租户分片库，只维护一份连接缓存。
	shards := newShardStore(db, c.Database.Source, c.GetHistory().GetDatabasePrefix(), logger)
	history := newHistoryStore(shards, c.GetHistory().GetTablePrefix())
	messageLogs := newMessageLogStore(shards, c.GetHistory().GetMessageTablePrefix())
	cleanup := func() {
		shards.close()
		if err := client.Close(); err != nil {
			logger.Error(fmt.Sprintf("close database: %v", err))
		}
	}
	deps := &Dependencies{db: db, client: client, box: box, history: history, messageLogs: messageLogs, shards: shards}
	// 集群客户端缓存按集群 UID 共享，集群仓储和资源仓储共用同一份。
	deps.clients = newClusterClients(deps, logger)
	return deps, cleanup, nil
}

func Migrate(ctx context.Context, source string) error {
	helper := slog.Default()
	db, err := openDB(source, helper)
	if err != nil {
		return err
	}
	defer db.Close()
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer client.Close()
	if err := ApplySchema(ctx, client, db); err != nil {
		return databaseErr(helper, err)
	}
	if err := seed(ctx, client, helper); err != nil {
		return err
	}
	return nil
}

func openDB(source string, helper *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", source)
	if err != nil {
		return nil, databaseErr(helper, err)
	}
	// 多副本部署时连接数按副本累加，必须显式设上限，否则 PostgreSQL 会被打满。
	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, databaseErr(helper, err)
	}
	return db, nil
}

const (
	// dbMaxOpenConns 是单副本的连接上限；多副本合计仍应小于 PostgreSQL 的 max_connections。
	dbMaxOpenConns    = 20
	dbMaxIdleConns    = 10
	dbConnMaxLifetime = 30 * time.Minute
	dbConnMaxIdleTime = 5 * time.Minute
)

func databaseErr(helper *slog.Logger, err error) error {
	if err == nil {
		return nil
	}
	helper.Error(fmt.Sprintf("database: %v", err))
	return merr.Server("database is unavailable")
}
