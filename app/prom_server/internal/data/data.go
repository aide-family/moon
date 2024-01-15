package data

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/pkg/conn"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData)

// Data .
type Data struct {
	db       *gorm.DB
	client   *redis.Client
	enforcer *casbin.SyncedEnforcer

	log *log.Helper
}

// DB gorm DB对象
func (d *Data) DB() *gorm.DB {
	return d.db
}

// Client redis client
func (d *Data) Client() *redis.Client {
	return d.client
}

// Enforcer casbin enforcer
func (d *Data) Enforcer() *casbin.SyncedEnforcer {
	return d.enforcer
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	databaseConf := c.GetData().GetDatabase()
	redisConf := c.GetData().GetRedis()
	env := c.GetEnv()
	db, err := conn.NewMysqlDB(databaseConf, logger)
	if err != nil {
		return nil, nil, err
	}
	d := &Data{
		log:    log.NewHelper(log.With(logger, "module", "data")),
		client: conn.NewRedisClient(redisConf),
		db:     db,
	}

	if env.GetEnv() == "dev" || env.GetEnv() == "test" {
		if err = do.Migrate(db); err != nil {
			d.log.Errorf("db migrate error: %v", err)
			return nil, nil, err
		}
	}

	if err = d.Client().Ping(context.Background()).Err(); err != nil {
		d.log.Errorf("redis ping error: %v", err)
		return nil, nil, err
	}

	d.enforcer, err = conn.InitCasbinModel(d.DB())
	if err != nil {
		d.log.Errorf("casbin init error: %v", err)
		return nil, nil, err
	}

	if err = do.CacheUserRoles(d.DB(), d.Client()); err != nil {
		return nil, nil, err
	}
	if err = do.CacheAllApiSimple(d.DB(), d.Client()); err != nil {
		return nil, nil, err
	}
	if err = do.CacheDisabledRoles(d.DB(), d.Client()); err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		sqlDb, err := d.DB().DB()
		if err != nil {
			log.NewHelper(logger).Errorf("close db error: %v", err)
		}
		if err = sqlDb.Close(); err != nil {
			log.NewHelper(logger).Errorf("close db error: %v", err)
		}
		log.NewHelper(logger).Info("closing the data resources")
	}

	return d, cleanup, nil
}
