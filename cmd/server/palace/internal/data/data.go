package data

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/conn/rbac"
	"github.com/aide-family/moon/pkg/util/email"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/casbin/casbin/v2"
	"github.com/coocood/freecache"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	bizDatabaseConf   *palaceconf.Data_Database
	alarmDatabaseConf *palaceconf.Data_Database

	mainDB       *gorm.DB
	alarmDB      *sql.DB
	bizDB        *sql.DB
	cacher       cache.ICacher
	enforcerMap  *sync.Map
	teamBizDBMap *sync.Map
	alarmDBMap   *sync.Map

	// 策略队列
	strategyQueue watch.Queue
	// 告警队列
	alertQueue watch.Queue
	// 持久化队列
	alertPersistenceDBQueue watch.Queue
	// 告警持久化存储
	alertConsumerStorage watch.Storage

	// 通用邮件发送器
	emailer email.Interface

	exit chan struct{}
}

var closeFuncList []func()

// NewData .
func NewData(c *palaceconf.Bootstrap) (*Data, func(), error) {
	mainConf := c.GetData().GetDatabase()
	alarmConf := c.GetData().GetAlarmDatabase()
	bizConf := c.GetData().GetBizDatabase()
	cacheConf := c.GetData().GetCache()
	emailConf := c.GetGlobalEmailConfig()
	d := &Data{
		bizDatabaseConf:         bizConf,
		alarmDatabaseConf:       alarmConf,
		teamBizDBMap:            new(sync.Map),
		alarmDBMap:              new(sync.Map),
		enforcerMap:             new(sync.Map),
		strategyQueue:           watch.NewDefaultQueue(100),
		alertQueue:              watch.NewDefaultQueue(100),
		alertPersistenceDBQueue: watch.NewDefaultQueue(100),
		alertConsumerStorage:    watch.NewDefaultStorage(),
		emailer:                 email.NewMockEmail(),
		exit:                    make(chan struct{}),
	}
	cleanup := func() {
		for _, f := range closeFuncList {
			f()
		}
		log.Info("closing the data resources")
	}
	closeFuncList = append(closeFuncList, func() {
		log.Debug("close data")
		d.exit <- struct{}{}
	})

	if !types.IsNil(emailConf) {
		emailer := email.New(emailConf)
		d.emailer = emailer
	}

	if !types.IsNil(cacheConf) {
		d.cacher = newCache(cacheConf)
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close cache", d.cacher.Close())
		})
	}

	if !types.IsNil(alarmConf) && !types.TextIsNull(alarmConf.GetDsn()) {
		// 打开数据库连接
		db, err := sql.Open(alarmConf.GetDriver(), bizConf.GetDsn())
		if !types.IsNil(err) {
			log.Fatalf("Error opening database: %v\n", err)
			cleanup()
			return nil, nil, err
		}
		d.alarmDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close alarm db", d.bizDB.Close())
			d.alarmDBMap.Range(func(key, value any) bool {
				alarmDB, ok := value.(*gorm.DB)
				if !ok {
					return true
				}
				dbTmp, _ := alarmDB.DB()
				log.Debugw("close alarm orm db", dbTmp.Close())
				return ok
			})
		})
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) {
		// 打开数据库连接
		db, err := sql.Open(bizConf.GetDriver(), bizConf.GetDsn())
		if !types.IsNil(err) {
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

	if !types.IsNil(mainConf) && !types.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn.NewGormDB(mainConf, log.GetLogger())
		if !types.IsNil(err) {
			cleanup()
			return nil, nil, err
		}
		d.mainDB = mainDB
		// 判断是否有默认用户， 如果没有，则创建一个默认用户
		if err := d.initMainDatabase(); err != nil {
			log.Fatalf("Error init default user: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		// 同步业务模型到各个团队， 保证数据一致性
		if err := d.syncBizDatabase(); err != nil {
			log.Fatalf("Error init biz database: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			log.Debugw("close main db", mainDBClose.Close())
		})
	}

	return d, cleanup, nil
}

// Exit 推出data
func (d *Data) Exit() <-chan struct{} {
	return d.exit
}

// initMainDatabase 初始化数据库
func (d *Data) initMainDatabase() error {
	return initMainDatabase(d)
}

// syncBizDatabase 同步业务模型到各个团队， 保证数据一致性
func (d *Data) syncBizDatabase() error {
	return syncBizDatabase(d)
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
func (d *Data) GetBizDB(_ context.Context) *sql.DB {
	return d.bizDB
}

// GetEmailer 获取邮件发送器
func (d *Data) GetEmailer() email.Interface {
	if types.IsNil(d.emailer) {
		return email.NewMockEmail()
	}
	return d.emailer
}

// GenBizDatabaseName 生成业务库名称
func GenBizDatabaseName(teamID uint32) string {
	return fmt.Sprintf("team_%d", teamID)
}

// GenAlarmBizDatabaseName 生成告警库名称
func GenAlarmBizDatabaseName(teamID uint32) string {
	return fmt.Sprintf("team_alarm_%d", teamID)
}

// GetBizGormDB 获取业务库连接
func (d *Data) GetBizGormDB(teamID uint32) (*gorm.DB, error) {
	if teamID == 0 {
		return nil, merr.ErrorI18nToastTeamNotFound(context.Background())
	}
	dbValue, exist := d.teamBizDBMap.Load(teamID)
	if exist {
		bizDB, isBizDB := dbValue.(*gorm.DB)
		if isBizDB {
			return bizDB, nil
		}
		return nil, merr.ErrorNotification("数据库服务异常")
	}

	dsn := d.bizDatabaseConf.GetDsn() + GenBizDatabaseName(teamID) + "?charset=utf8mb4&parseTime=True&loc=Local"
	bizDbConf := &palaceconf.Data_Database{
		Driver: d.bizDatabaseConf.GetDriver(),
		Dsn:    dsn,
		Debug:  d.bizDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(bizDbConf, log.GetLogger())
	if !types.IsNil(err) {
		return nil, err
	}

	d.teamBizDBMap.Store(teamID, bizDB)
	enforcer, err := rbac.InitCasbinModel(bizDB)
	if !types.IsNil(err) {
		log.Errorw("casbin init error", err)
		return nil, err
	}
	d.enforcerMap.Store(teamID, enforcer)
	return bizDB, nil
}

// GetAlarmGormDB 获取告警库连接
func (d *Data) GetAlarmGormDB(teamID uint32) (*gorm.DB, error) {
	if teamID == 0 {
		return nil, merr.ErrorI18nToastTeamNotFound(context.Background())
	}
	dbValue, exist := d.alarmDBMap.Load(teamID)
	if exist {
		bizDB, isBizDB := dbValue.(*gorm.DB)
		if isBizDB {
			return bizDB, nil
		}
		return nil, merr.ErrorNotification("数据库服务异常")
	}

	dsn := d.alarmDatabaseConf.GetDsn() + GenAlarmBizDatabaseName(teamID) + "?charset=utf8mb4&parseTime=True&loc=Local"
	alarmDbConf := &palaceconf.Data_Database{
		Driver: d.alarmDatabaseConf.GetDriver(),
		Dsn:    dsn,
		Debug:  d.alarmDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(alarmDbConf, log.GetLogger())
	if !types.IsNil(err) {
		return nil, err
	}

	d.alarmDBMap.Store(teamID, bizDB)
	return bizDB, nil
}

// GetCacher 获取缓存
func (d *Data) GetCacher() cache.ICacher {
	if types.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// GetCasbin 获取casbin
func (d *Data) GetCasbin(teamID uint32) *casbin.SyncedEnforcer {
	enforceVal, exist := d.enforcerMap.Load(teamID)
	if !exist {
		_, err := d.GetBizGormDB(teamID)
		if !types.IsNil(err) {
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
func newCache(c *palaceconf.Data_Cache) cache.ICacher {
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

	size := int(c.GetFree().GetSize())
	return cache.NewFreeCache(freecache.NewCache(types.Ternary(size > 0, size, 10*1024*1024)))
}

// GetStrategyQueue 获取策略队列
func (d *Data) GetStrategyQueue() watch.Queue {
	if types.IsNil(d.strategyQueue) {
		log.Warn("strategyQueue is nil")
	}
	return d.strategyQueue
}

// GetAlertQueue 获取告警队列
func (d *Data) GetAlertQueue() watch.Queue {
	if types.IsNil(d.alertQueue) {
		log.Warn("alertQueue is nil")
	}
	return d.alertQueue
}

// GetAlertPersistenceDBQueue 获取持久化队列
func (d *Data) GetAlertPersistenceDBQueue() watch.Queue {
	if types.IsNil(d.alertPersistenceDBQueue) {
		log.Warn("persistence queue is nil")
	}
	return d.alertPersistenceDBQueue
}

// GetAlertConsumerStorage 获取告警持久化存储
func (d *Data) GetAlertConsumerStorage() watch.Storage {
	if types.IsNil(d.alertConsumerStorage) {
		log.Warn("alertConsumerStorage is nil")
	}
	return d.alertConsumerStorage
}
