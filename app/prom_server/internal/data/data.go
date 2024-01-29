package data

import (
	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/util/cache"

	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/pkg/conn"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(
	NewData,
	GetReadChangeGroupChannel,
	GetWriteChangeGroupChannel,
	GetReadRemoveGroupChannel,
	GetWriteRemoveGroupChannel,
)

var changeGroupChannel = make(chan uint32, 100)
var removeGroupChannel = make(chan bo.RemoveStrategyGroupBO, 100)

// GetReadChangeGroupChannel 获取changeGroupChannel的读取通道
func GetReadChangeGroupChannel() <-chan uint32 {
	return changeGroupChannel
}

// GetWriteChangeGroupChannel 获取changeGroupChannel的写入通道
func GetWriteChangeGroupChannel() chan<- uint32 {
	return changeGroupChannel
}

// GetReadRemoveGroupChannel 获取removeGroupChannel的读取通道
func GetReadRemoveGroupChannel() <-chan bo.RemoveStrategyGroupBO {
	return removeGroupChannel
}

// GetWriteRemoveGroupChannel 获取removeGroupChannel的写入通道
func GetWriteRemoveGroupChannel() chan<- bo.RemoveStrategyGroupBO {
	return removeGroupChannel
}

// Data .
type Data struct {
	db *gorm.DB
	//client   *redis.Client
	cache    cache.GlobalCache
	enforcer *casbin.SyncedEnforcer

	log *log.Helper
}

// DB gorm DB对象
func (d *Data) DB() *gorm.DB {
	return d.db
}

// Cache cache
func (d *Data) Cache() cache.GlobalCache {
	return d.cache
}

// Enforcer casbin enforcer
func (d *Data) Enforcer() *casbin.SyncedEnforcer {
	return d.enforcer
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	databaseConf := c.GetData().GetDatabase()

	env := c.GetEnv()
	db, err := conn.NewMysqlDB(databaseConf, logger)
	if err != nil {
		return nil, nil, err
	}
	//redisConf := c.GetData().GetRedis()
	//globalCache := cache.NewRedisGlobalCache(conn.NewRedisClient(redisConf))
	globalCache, err := cache.NewNutsDbCache()
	if err != nil {
		return nil, nil, err
	}
	d := &Data{
		log:   log.NewHelper(log.With(logger, "module", "data")),
		db:    db,
		cache: globalCache,
	}

	if env.GetEnv() == "dev" || env.GetEnv() == "test" {
		if err = do.Migrate(db); err != nil {
			d.log.Errorf("db migrate error: %v", err)
			return nil, nil, err
		}
	}

	d.enforcer, err = conn.InitCasbinModel(d.DB())
	if err != nil {
		d.log.Errorf("casbin init error: %v", err)
		return nil, nil, err
	}

	if err = do.CacheUserRoles(d.DB(), d.Cache()); err != nil {
		d.log.Errorf("cache user roles error: %v", err)
		return nil, nil, err
	}
	if err = do.CacheAllApiSimple(d.DB(), d.Cache()); err != nil {
		d.log.Errorf("cache all api simple error: %v", err)
		return nil, nil, err
	}
	if err = do.CacheDisabledRoles(d.DB(), d.Cache()); err != nil {
		d.log.Errorf("cache disabled roles error: %v", err)
		return nil, nil, err
	}

	cleanup := func() {
		sqlDb, err := d.DB().DB()
		if err != nil {
			d.log.Errorf("close db error: %v", err)
		}
		if err = sqlDb.Close(); err != nil {
			d.log.Errorf("close db error: %v", err)
		}
		if err = globalCache.Close(); err != nil {
			d.log.Errorf("close cache error: %v", err)
		}
		d.log.Info("closing the data resources")
	}

	return d, cleanup, nil
}
