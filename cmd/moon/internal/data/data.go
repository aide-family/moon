package data

import (
	"context"

	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/conn/cacher/nutsdbcacher"
	"github.com/aide-cloud/moon/pkg/conn/cacher/rediscacher"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/moon/internal/conf"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	mainDB *gorm.DB
	bizDB  *gorm.DB
	cacher conn.Cache
}

var closeFuncList []func()

// NewData .
func NewData(c *conf.Bootstrap) (*Data, func(), error) {
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
		mainDB, err := conn.NewGormDB(mainConf.GetDsn(), mainConf.GetDriver())
		if err != nil {
			return nil, nil, err
		}
		d.mainDB = mainDB
		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			log.Debugw("close main db", mainDBClose.Close())
		})
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) {
		bizDB, err := conn.NewGormDB(bizConf.GetDsn(), bizConf.GetDriver())
		if err != nil {
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
	db, exist := ctx.Value(conn.GormContextTxKey{}).(*gorm.DB)
	if exist {
		return db
	}
	return d.mainDB
}

// GetBizDB 获取业务库连接
func (d *Data) GetBizDB(ctx context.Context) *gorm.DB {
	db, exist := ctx.Value(conn.GormContextTxKey{}).(*gorm.DB)
	if exist {
		return db
	}
	return d.bizDB
}

// GetCacher 获取缓存
func (d *Data) GetCacher() conn.Cache {
	if types.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// newCache new cache
func newCache(c *conf.Data_Cache) conn.Cache {
	if types.IsNil(c) {
		return nil
	}
	switch {
	case !types.IsNil(c.GetRedis()):
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); err != nil {
			log.Warnw("redis ping error", err)
		}
		return rediscacher.NewRedisCacher(cli)
	case !types.IsNil(c.GetNutsDB()):
		cli, err := nutsdbcacher.NewNutsDbCacher(c.GetNutsDB())
		if err != nil {
			log.Warnw("nutsdb init error", err)
		}
		return cli
	default:
		return nil
	}
}
