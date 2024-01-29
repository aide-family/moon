package data

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/pkg/conn"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/cache"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewPingRepo)

// Data .
type Data struct {
	storeDB    *gorm.DB
	cache      cache.GlobalCache
	mqProducer *kafka.Producer

	log *log.Helper
}

func (d *Data) StoreDB() *gorm.DB {
	return d.storeDB
}

func (d *Data) Cache() cache.GlobalCache {
	return d.cache
}

func (d *Data) Producer() *kafka.Producer {
	return d.mqProducer
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	databaseConf := c.GetData().GetDatabase()
	mqConf := c.GetMq().GetKafka()
	db, err := conn.NewMysqlDB(databaseConf, logger)
	if err != nil {
		return nil, nil, err
	}
	kafkaProducer, err := conn.NewKafkaProducer(mqConf.GetEndpoints())
	if err != nil {
		return nil, nil, err
	}
	//redisConf := c.GetData().GetRedis()
	//globalCache :=cache.NewRedisGlobalCache(conn.NewRedisClient(redisConf))
	globalCache, err := cache.NewNutsDbCache()
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		log:        log.NewHelper(log.With(logger, "module", "data")),
		cache:      globalCache,
		storeDB:    db,
		mqProducer: kafkaProducer,
	}

	// 注册全局告警缓存组件
	alarmCache := strategy.NewAlarmCache(d.Cache())
	strategy.SetAlarmCache(alarmCache)

	cleanup := func() {
		sqlDb, err := d.StoreDB().DB()
		if err != nil {
			d.log.Errorf("close db error: %v", err)
		}
		if err = sqlDb.Close(); err != nil {
			d.log.Errorf("close db error: %v", err)
		}
		if err = d.Cache().Close(); err != nil {
			d.log.Errorf("close redis error: %v", err)
		}
		d.mqProducer.Close()
		d.log.Info("closing the data resources")
	}
	return d, cleanup, nil
}
