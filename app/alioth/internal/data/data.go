package data

import (
	_ "github.com/jackc/pgx/v5/stdlib"

	"context"
	"database/sql"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
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
)

type Data struct {
	db     *sql.DB
	client *ent.Client
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	if c == nil || c.Database == nil || c.Database.Source == "" {
		return nil, nil, merr.Server("database source is empty")
	}
	helper := log.NewHelper(logger)
	db, err := openDB(c.Database.Source, helper)
	if err != nil {
		return nil, nil, err
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	cleanup := func() {
		if err := client.Close(); err != nil {
			log.NewHelper(logger).Errorf("close database: %v", err)
		}
	}
	return &Data{db: db, client: client}, cleanup, nil
}

func Migrate(ctx context.Context, source string) error {
	helper := log.NewHelper(log.DefaultLogger)
	db, err := openDB(source, helper)
	if err != nil {
		return err
	}
	defer db.Close()
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		return databaseErr(helper, err)
	}
	if err := seed(ctx, client, helper); err != nil {
		return err
	}
	return nil
}

func openDB(source string, helper *log.Helper) (*sql.DB, error) {
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

func databaseErr(helper *log.Helper, err error) error {
	if err == nil {
		return nil
	}
	helper.Errorf("database: %v", err)
	return merr.Server("database is unavailable")
}
