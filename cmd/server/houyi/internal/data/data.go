package data

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	conn2 "github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/conn/cacher/nutsdbcacher"
	"github.com/aide-family/moon/pkg/util/conn/cacher/rediscacher"
	types2 "github.com/aide-family/moon/pkg/util/types"

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
	cacher   conn2.Cache
	enforcer *casbin.SyncedEnforcer
}

var closeFuncList []func()

// NewData .
func NewData(c *houyiconf.Bootstrap) (*Data, func(), error) {
	d := &Data{}
	mainConf := c.GetData().GetDatabase()
	bizConf := c.GetData().GetBizDatabase()
	cacheConf := c.GetData().GetCache()
	if !types2.IsNil(cacheConf) {
		d.cacher = newCache(cacheConf)
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close cache", d.cacher.Close())
		})
	}

	if !types2.IsNil(mainConf) && !types2.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn2.NewGormDB(mainConf.GetDsn(), mainConf.GetDriver())
		if !types2.IsNil(err) {
			return nil, nil, err
		}
		d.mainDB = mainDB
		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			log.Debugw("close main db", mainDBClose.Close())
		})
		// 开发需要开启
		//query.SetDefault(mainDB)
	}

	if !types2.IsNil(bizConf) && !types2.TextIsNull(bizConf.GetDsn()) {
		bizDB, err := conn2.NewGormDB(bizConf.GetDsn(), bizConf.GetDriver())
		if !types2.IsNil(err) {
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
	db, exist := ctx.Value(conn2.GormContextTxKey{}).(*gorm.DB)
	if exist {
		return db
	}
	return d.mainDB
}

// GetBizDB 获取业务库连接
func (d *Data) GetBizDB(ctx context.Context) *gorm.DB {
	db, exist := ctx.Value(conn2.GormContextTxKey{}).(*gorm.DB)
	if exist {
		return db
	}
	return d.bizDB
}

// GetCacher 获取缓存
func (d *Data) GetCacher() conn2.Cache {
	if types2.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// GetCasbin 获取casbin
func (d *Data) GetCasbin() *casbin.SyncedEnforcer {
	return d.enforcer
}

// newCache new cache
func newCache(c *houyiconf.Data_Cache) conn2.Cache {
	if types2.IsNil(c) {
		return nil
	}

	if !types2.IsNil(c.GetRedis()) {
		log.Debugw("cache init", "redis")
		cli := conn2.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types2.IsNil(err) {
			log.Warnw("redis ping error", err)
		}
		return rediscacher.NewRedisCacher(cli)
	}

	if !types2.IsNil(c.GetNutsDB()) {
		log.Debugw("cache init", "nutsdb")
		cli, err := nutsdbcacher.NewNutsDbCacher(c.GetNutsDB())
		if !types2.IsNil(err) {
			log.Warnw("nutsdb init error", err)
		}
		return cli
	}
	return nil
}
