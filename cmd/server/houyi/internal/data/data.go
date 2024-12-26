package data

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/coocood/freecache"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData)

// Data .
type Data struct {
	cacher cache.ICacher

	strategyQueue      watch.Queue
	eventMQQueue       watch.Queue
	eventStrategyQueue watch.Queue

	alertQueue   watch.Queue
	alertStorage watch.Storage
}

var closeFuncList []func()

// NewData .
func NewData(c *houyiconf.Bootstrap) (*Data, func(), error) {
	d := &Data{
		strategyQueue:      watch.NewDefaultQueue(watch.QueueMaxSize),
		eventMQQueue:       watch.NewDefaultQueue(watch.QueueMaxSize),
		eventStrategyQueue: watch.NewDefaultQueue(watch.QueueMaxSize),
		alertQueue:         watch.NewDefaultQueue(watch.QueueMaxSize),
		alertStorage:       watch.NewDefaultStorage(),
	}

	cacheConf := c.GetCache()
	d.cacher = newCache(cacheConf)
	d.alertStorage = watch.NewCacheStorage(d.cacher)
	closeFuncList = append(closeFuncList, func() {
		log.Debugw("close alert storage", d.alertStorage.Close())
		log.Debugw("close cache", d.cacher.Close())
	})

	cleanup := func() {
		for _, f := range closeFuncList {
			f()
		}
		log.Info("closing the data resources")
	}
	return d, cleanup, nil
}

// GetCacher 获取缓存
func (d *Data) GetCacher() cache.ICacher {
	if types.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
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

// GetAlertStorage 获取告警存储
func (d *Data) GetAlertStorage() watch.Storage {
	if types.IsNil(d.alertStorage) {
		log.Warn("alertStorage is nil")
	}
	return d.alertStorage
}

// GetEventMQQueue 获取事件队列
func (d *Data) GetEventMQQueue() watch.Queue {
	if types.IsNil(d.eventMQQueue) {
		log.Warn("eventMQQueue is nil")
	}
	return d.eventMQQueue
}

// GetEventStrategyQueue 获取事件策略队列
func (d *Data) GetEventStrategyQueue() watch.Queue {
	if types.IsNil(d.eventStrategyQueue) {
		log.Warn("eventStrategyQueue is nil")
	}
	return d.eventStrategyQueue
}

// newCache new cache
func newCache(c *conf.Cache) cache.ICacher {
	switch c.GetDriver() {
	case "redis", "REDIS":
		log.Debugw("cache init", "redis")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types.IsNil(err) {
			log.Warnw("redis ping error", err)
		}
		return cache.NewRedisCacher(cli)
	default:
		log.Debugw("cache init", "free")
		size := int(c.GetFree().GetSize())
		return cache.NewFreeCache(freecache.NewCache(types.Ternary(size > 0, size, 10*1024*1024)))
	}
}
