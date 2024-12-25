package data

import (
	"context"

	"github.com/aide-family/moon/cmd/server/demo/internal/democonf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/coocood/freecache"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	mainDB   *gorm.DB
	bizDB    *gorm.DB
	cacher   cache.ICacher
	enforcer *casbin.SyncedEnforcer
}

var closeFuncList []func()

// NewData .
func NewData(c *democonf.Bootstrap) (*Data, func(), error) {
	d := &Data{}
	mainConf := c.GetData().GetDatabase()
	bizConf := c.GetData().GetBizDatabase()
	cacheConf := c.GetData().GetCache()
	if !types.IsNil(cacheConf) {
		d.cacher = newCache(cacheConf)
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close cache", d.cacher.Close())
		})
	}

	if !types.IsNil(mainConf) && !types.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn.NewGormDB(mainConf)
		if !types.IsNil(err) {
			return nil, nil, err
		}
		d.mainDB = mainDB
		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			log.Debugw("close main db", mainDBClose.Close())
		})
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) {
		bizDB, err := conn.NewGormDB(bizConf)
		if !types.IsNil(err) {
			return nil, nil, err
		}
		d.bizDB = bizDB
		closeFuncList = append(closeFuncList, func() {
			bizDBClose, _ := d.bizDB.DB()
			log.Debugw("close biz db", bizDBClose.Close())
		})
	}

	cleanup := func() {
		for _, f := range closeFuncList {
			f()
		}
		log.Info("closing the data resources")
	}
	return d, cleanup, nil
}

// GetMainDB 获取主库连接
func (d *Data) GetMainDB(ctx context.Context) *gorm.DB {
	db, exist := conn.GetDB(ctx)
	if exist {
		return db
	}
	return d.mainDB
}

// GetBizDB 获取业务库连接
func (d *Data) GetBizDB(ctx context.Context) *gorm.DB {
	db, exist := conn.GetDB(ctx)
	if exist {
		return db
	}
	return d.bizDB
}

// GetCacher 获取缓存
func (d *Data) GetCacher() cache.ICacher {
	if types.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// GetCasbin 获取casbin
func (d *Data) GetCasbin() *casbin.SyncedEnforcer {
	return d.enforcer
}

// newCache new cache
func newCache(c *democonf.Data_Cache) cache.ICacher {
	if types.IsNil(c) {
		return nil
	}

	if !types.IsNil(c.GetRedis()) {
		log.Debugw("cache init", "redis")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types.IsNil(err) {
			log.Warnw("redis ping error", err)
		}
		return cache.NewRedisCacher(cli)
	}

	if !types.IsNil(c.GetNutsDB()) {
		log.Debugw("cache init", "nutsdb")
		cli, err := conn.NewNutsDB(c.GetNutsDB())
		if !types.IsNil(err) {
			log.Warnw("nutsdb init error", err)
		}
		return cache.NewNutsDbCacher(cli, c.GetNutsDB().GetBucket())
	}

	return cache.NewFreeCache(freecache.NewCache(10 * 1024 * 1024))
}
