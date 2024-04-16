package data

import (
	"time"

	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/aide-family/moon/pkg/helper"
	"github.com/aide-family/moon/pkg/servers"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/cache"
	"github.com/aide-family/moon/pkg/util/email"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/aide-family/moon/pkg/conn"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(
	NewData,
)

// Data .
type Data struct {
	db                *gorm.DB
	cache             cache.GlobalCache
	enforcer          *casbin.SyncedEnforcer
	interflowInstance interflow.Interflow
	email             email.Interface

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

func (d *Data) Interflow() interflow.Interflow {
	return d.interflowInstance
}

// Enforcer casbin enforcer
func (d *Data) Enforcer() *casbin.SyncedEnforcer {
	return d.enforcer
}

// Email email
func (d *Data) Email() email.Interface {
	return d.email
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	databaseConf := c.GetData().GetDatabase()

	env := c.GetEnv()
	var logList []log.Logger
	if !helper.IsDev(env.GetEnv()) {
		// 其他环境使用系统日志，开发环境使用原始调试日志
		logList = []log.Logger{logger}
	}
	db, err := conn.NewDB(databaseConf, logList...)
	if err != nil {
		return nil, nil, err
	}
	// 设置数据库连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	d := &Data{
		log: log.NewHelper(log.With(logger, "module", "data")),
		db:  db,
	}

	if c.GetEmail() != nil {
		d.email = email.New(c.GetEmail())
	}

	redisConf := c.GetData().GetRedis()
	if redisConf != nil {
		d.cache = cache.NewRedisGlobalCache(conn.NewRedisClient(redisConf))
	} else {
		globalCache, err := cache.NewNutsDbCache()
		if err != nil {
			return nil, nil, err
		}
		d.cache = globalCache
	}
	// 注册全局告警缓存组件
	alarmCache := strategy.NewAlarmCache(d.Cache())
	strategy.SetAlarmCache(alarmCache)

	kafkaConf := c.GetMq().GetKafka()
	if kafkaConf != nil {
		kafkaMqServer, err := servers.NewKafkaMQServer(kafkaConf, logger)
		if err != nil {
			return nil, nil, err
		}
		interflowInstance, err := interflow.NewKafkaInterflow(kafkaMqServer, d.log)
		if err != nil {
			d.log.Errorf("init kafka interflow error: %v", err)
			return nil, nil, err
		}
		d.interflowInstance = interflowInstance
	} else {
		interflowInstance := interflow.NewHookInterflow(d.log)
		d.interflowInstance = interflowInstance
	}

	// TODO casbin config
	//d.enforcer, err = conn.InitCasbinModel(d.DB())
	//if err != nil {
	//	d.log.Errorf("casbin init error: %v", err)
	//	return nil, nil, err
	//}

	cleanup := func() {
		sqlDb, err := d.DB().DB()
		if err != nil {
			d.log.Errorf("close db error: %v", err)
		}
		if err = sqlDb.Close(); err != nil {
			d.log.Errorf("close db error: %v", err)
		}
		if err = d.cache.Close(); err != nil {
			d.log.Errorf("close cache error: %v", err)
		}
		if err = d.Interflow().Close(); err != nil {
			d.log.Errorf("close interflow error: %v", err)
		}

		d.log.Info("closing the data resources")
	}

	return d, cleanup, nil
}
