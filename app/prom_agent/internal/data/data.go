package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/pkg/conn"
	"prometheus-manager/pkg/servers"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/cache"
	"prometheus-manager/pkg/util/interflow"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewPingRepo)

// Data .
type Data struct {
	cache             cache.GlobalCache
	interflowInstance interflow.Interflow

	log *log.Helper
}

func (d *Data) Cache() cache.GlobalCache {
	return d.cache
}

func (d *Data) Interflow() interflow.Interflow {
	return d.interflowInstance
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	d := &Data{
		log: log.NewHelper(log.With(logger, "module", "data")),
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

	kafkaConf := c.GetMq().GetKafka()
	if kafkaConf != nil {
		kafkaMqServer, err := servers.NewKafkaMQServer(kafkaConf, logger)
		if err != nil {
			return nil, nil, err
		}
		interflowInstance, err := interflow.NewKafkaInterflow(kafkaMqServer, d.log)
		if err != nil {
			return nil, nil, err
		}
		d.interflowInstance = interflowInstance
	} else {
		interflowInstance := interflow.NewHookInterflow(d.log)
		d.interflowInstance = interflowInstance
	}

	// 注册全局告警缓存组件
	alarmCache := strategy.NewAlarmCache(d.Cache())
	strategy.SetAlarmCache(alarmCache)

	cleanup := func() {
		if err := d.Cache().Close(); err != nil {
			d.log.Errorf("close redis error: %v", err)
		}
		if err := d.Interflow().Close(); err != nil {
			d.log.Errorf("close interflow error: %v", err)
		}
		d.log.Info("closing the data resources")
	}
	return d, cleanup, nil
}
