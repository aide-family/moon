package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/conn/cacher/nutsdbcacher"
	"github.com/aide-cloud/moon/pkg/conn/cacher/rediscacher"
	"github.com/aide-cloud/moon/pkg/conn/rbac"
	"github.com/aide-cloud/moon/pkg/helper/model/query"
	"github.com/aide-cloud/moon/pkg/types"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	mainDB          *gorm.DB
	bizDB           *sql.DB
	cacher          conn.Cache
	enforcer        *casbin.SyncedEnforcer
	bizDatabaseConf *palaceconf.Data_Database
	teamBizDBMap    map[uint32]*gorm.DB
}

var closeFuncList []func()

// NewData .
func NewData(c *palaceconf.Bootstrap) (*Data, func(), error) {
	mainConf := c.GetData().GetDatabase()
	bizConf := c.GetData().GetBizDatabase()
	cacheConf := c.GetData().GetCache()
	d := &Data{
		bizDatabaseConf: bizConf,
		teamBizDBMap:    make(map[uint32]*gorm.DB),
	}

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
		query.SetDefault(d.mainDB)
		d.enforcer, err = rbac.InitCasbinModel(d.mainDB)
		if err != nil {
			log.Errorw("casbin init error", err)
			return nil, nil, err
		}
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) {
		// 打开数据库连接
		db, err := sql.Open(bizConf.GetDriver(), bizConf.GetDsn())
		if err != nil {
			log.Fatalf("Error opening database: %v\n", err)
		}

		d.bizDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close biz db", d.bizDB.Close())
			for _, db := range d.teamBizDBMap {
				dbTmp, _ := db.DB()
				log.Debugw("close biz orm db", dbTmp.Close())
			}
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
func (d *Data) GetBizDB(ctx context.Context) *sql.DB {
	db, exist := ctx.Value(conn.GormContextTxKey{}).(*sql.DB)
	if exist {
		return db
	}
	return d.bizDB
}

// GenBizDatabaseName 生成业务库名称
func GenBizDatabaseName(teamId uint32) string {
	return fmt.Sprintf("team_%d", teamId)
}

// GetBizGormDB 获取业务库连接
func (d *Data) GetBizGormDB(teamId uint32) (*gorm.DB, error) {
	bizDB, exist := d.teamBizDBMap[teamId]
	if exist {
		return bizDB, nil
	}

	dsn := d.bizDatabaseConf.GetDsn() + GenBizDatabaseName(teamId) + "?charset=utf8mb4&parseTime=True&loc=Local"
	bizDB, err := conn.NewGormDB(dsn, d.bizDatabaseConf.GetDriver())
	if err != nil {
		return nil, err
	}
	d.teamBizDBMap[teamId] = bizDB

	return bizDB, nil
}

// GetCacher 获取缓存
func (d *Data) GetCacher() conn.Cache {
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
func newCache(c *palaceconf.Data_Cache) conn.Cache {
	if types.IsNil(c) {
		return nil
	}

	if !types.IsNil(c.GetRedis()) {
		log.Debugw("cache init", "redis")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); err != nil {
			log.Warnw("redis ping error", err)
		}
		return rediscacher.NewRedisCacher(cli)
	}

	if !types.IsNil(c.GetNutsDB()) {
		log.Debugw("cache init", "nutsdb")
		cli, err := nutsdbcacher.NewNutsDbCacher(c.GetNutsDB())
		if err != nil {
			log.Warnw("nutsdb init error", err)
		}
		return cli
	}
	return nil
}
