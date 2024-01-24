package data

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"prometheus-manager/pkg/conn"
	"prometheus-manager/pkg/strategy"

	"prometheus-manager/app/prom_agent/internal/conf"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewPingRepo)

// Data .
type Data struct {
	storeDB     *gorm.DB
	cacheClient *redis.Client
	mqProducer  *kafka.Producer

	log *log.Helper
}

func (d *Data) StoreDB() *gorm.DB {
	return d.storeDB
}

func (d *Data) CacheClient() *redis.Client {
	return d.cacheClient
}

func (d *Data) Producer() *kafka.Producer {
	return d.mqProducer
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	databaseConf := c.GetData().GetDatabase()
	redisConf := c.GetData().GetRedis()
	mqConf := c.GetMq().GetKafka()
	db, err := conn.NewMysqlDB(databaseConf, logger)
	if err != nil {
		return nil, nil, err
	}
	kafkaProducer, err := conn.NewKafkaProducer(mqConf.GetEndpoints())
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		log:         log.NewHelper(log.With(logger, "module", "data")),
		cacheClient: conn.NewRedisClient(redisConf),
		storeDB:     db,
		mqProducer:  kafkaProducer,
	}
	// 注册全局告警缓存组件
	alarmCache := strategy.NewRedisAlarmCache(d.CacheClient())
	strategy.SetAlarmCache(alarmCache)

	cleanup := func() {
		sqlDb, err := d.StoreDB().DB()
		if err != nil {
			log.NewHelper(logger).Errorf("close db error: %v", err)
		}
		if err = sqlDb.Close(); err != nil {
			log.NewHelper(logger).Errorf("close db error: %v", err)
		}
		if err = d.CacheClient().Close(); err != nil {
			log.NewHelper(logger).Errorf("close redis error: %v", err)
		}
		d.mqProducer.Close()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return d, cleanup, nil
}
