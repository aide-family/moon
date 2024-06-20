package data

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/palace/model/query"
	conn2 "github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/conn/cacher/nutsdbcacher"
	"github.com/aide-family/moon/pkg/util/conn/cacher/rediscacher"
	"github.com/aide-family/moon/pkg/util/conn/rbac"
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
	bizDatabaseConf *palaceconf.Data_Database

	mainDB       *gorm.DB
	bizDB        *sql.DB
	cacher       conn2.Cache
	enforcerMap  *sync.Map
	teamBizDBMap *sync.Map

	exit chan struct{}
}

var closeFuncList []func()

// NewData .
func NewData(c *palaceconf.Bootstrap) (*Data, func(), error) {
	mainConf := c.GetData().GetDatabase()
	bizConf := c.GetData().GetBizDatabase()
	cacheConf := c.GetData().GetCache()
	d := &Data{
		bizDatabaseConf: bizConf,
		teamBizDBMap:    new(sync.Map),
		enforcerMap:     new(sync.Map),
		exit:            make(chan struct{}),
	}
	cleanup := func() {
		for _, f := range closeFuncList {
			f()
		}
		log.Info("closing the data resources")
	}
	closeFuncList = append(closeFuncList, func() {
		log.Debugw("close data")
		d.exit <- struct{}{}
	})

	if !types2.IsNil(cacheConf) {
		d.cacher = newCache(cacheConf)
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close cache", d.cacher.Close())
		})
	}

	if !types2.IsNil(mainConf) && !types2.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn2.NewGormDB(mainConf.GetDsn(), mainConf.GetDriver())
		if !types2.IsNil(err) {
			cleanup()
			return nil, nil, err
		}
		d.mainDB = mainDB
		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			log.Debugw("close main db", mainDBClose.Close())
		})
		query.SetDefault(d.mainDB)
	}

	if !types2.IsNil(bizConf) && !types2.TextIsNull(bizConf.GetDsn()) {
		// 打开数据库连接
		db, err := sql.Open(bizConf.GetDriver(), bizConf.GetDsn())
		if !types2.IsNil(err) {
			log.Fatalf("Error opening database: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		d.bizDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close biz db", d.bizDB.Close())
			d.teamBizDBMap.Range(func(key, value any) bool {
				teamBizDB, ok := value.(*gorm.DB)
				if !ok {
					return true
				}
				dbTmp, _ := teamBizDB.DB()
				log.Debugw("close biz orm db", dbTmp.Close())
				return ok
			})
		})
	}

	return d, cleanup, nil
}

func (d *Data) Exit() <-chan struct{} {
	return d.exit
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
func (d *Data) GetBizDB(ctx context.Context) *sql.DB {
	db, exist := ctx.Value(conn2.GormContextTxKey{}).(*sql.DB)
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
	if teamId == 0 {
		return nil, merr.ErrorI18nNoTeamErr(context.Background())
	}
	dbValue, exist := d.teamBizDBMap.Load(teamId)
	if exist {
		bizDB, isBizDB := dbValue.(*gorm.DB)
		if isBizDB {
			return bizDB, nil
		}
		return nil, merr.ErrorNotification("数据库服务异常")
	}

	dsn := d.bizDatabaseConf.GetDsn() + GenBizDatabaseName(teamId) + "?charset=utf8mb4&parseTime=True&loc=Local"
	bizDB, err := conn2.NewGormDB(dsn, d.bizDatabaseConf.GetDriver())
	if !types2.IsNil(err) {
		return nil, err
	}

	d.teamBizDBMap.Store(teamId, bizDB)
	enforcer, err := rbac.InitCasbinModel(bizDB)
	if !types2.IsNil(err) {
		log.Errorw("casbin init error", err)
		return nil, err
	}
	d.enforcerMap.Store(teamId, enforcer)
	return bizDB, nil
}

// GetCacher 获取缓存
func (d *Data) GetCacher() conn2.Cache {
	if types2.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// GetCasbin 获取casbin
func (d *Data) GetCasbin(teamId uint32) *casbin.SyncedEnforcer {
	enforceVal, exist := d.enforcerMap.Load(teamId)
	if !exist {
		_, err := d.GetBizGormDB(teamId)
		if !types2.IsNil(err) {
			panic(err)
		}
	}
	enforce, ok := enforceVal.(*casbin.SyncedEnforcer)
	if !ok {
		panic("enforcer not found")
	}
	return enforce
}

// newCache new cache
func newCache(c *palaceconf.Data_Cache) conn2.Cache {
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
