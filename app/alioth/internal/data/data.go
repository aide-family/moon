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

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(
	NewData,
	NewUserRepo,
	NewCaptchaRepo,
	NewEmailCodeRepo,
	NewOAuthRepo,
	NewNamespaceRepo,
	NewMemberRepo,
	NewRoleRepo,
	NewDatasourceRepo,
	NewAlertDictionaryRepo,
	NewAlertRuleRepo,
	NewAlertEventRepo,
	NewAlertPageRepo,
	NewNotificationRepo,
	NewAgentRepo,
	NewSSHCommandRepo,
	NewEdgeAuditRepo,
	NewLock,
)

type Data struct {
	db     *sql.DB
	client *ent.Client
	box    *secretBox
}

func NewData(c *conf.Data, auth *conf.Auth, logger *slog.Logger) (*Data, func(), error) {
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
	cleanup := func() {
		if err := client.Close(); err != nil {
			logger.Error(fmt.Sprintf("close database: %v", err))
		}
	}
	return &Data{db: db, client: client, box: box}, cleanup, nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, databaseErr(helper, err)
	}
	return db, nil
}

func databaseErr(helper *slog.Logger, err error) error {
	if err == nil {
		return nil
	}
	helper.Error(fmt.Sprintf("database: %v", err))
	return merr.Server("database is unavailable")
}
