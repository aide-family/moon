package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ProviderSet = wire.NewSet(NewData)

type Data struct {
	db *sql.DB
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	if c == nil || c.Database == nil || c.Database.Source == "" {
		return nil, nil, fmt.Errorf("database source is empty")
	}
	db, err := openDB(c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		if err := db.Close(); err != nil {
			log.NewHelper(logger).Errorf("close database: %v", err)
		}
	}
	return &Data{db: db}, cleanup, nil
}

func Migrate(ctx context.Context, source string) error {
	db, err := openDB(source)
	if err != nil {
		return err
	}
	defer db.Close()
	drv := entsql.OpenDB(dialect.Postgres, db)
	migrator, err := schema.NewMigrate(drv)
	if err != nil {
		return err
	}
	return migrator.Create(ctx)
}

func openDB(source string) (*sql.DB, error) {
	db, err := sql.Open("pgx", source)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
