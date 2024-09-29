package data

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/coocood/freecache"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	cacher cache.ICacher

	strategyQueue watch.Queue

	alertQueue   watch.Queue
	alertStorage watch.Storage
}

var closeFuncList []func()

// NewData .
func NewData(c *houyiconf.Bootstrap) (*Data, func(), error) {
	d := &Data{
		strategyQueue: watch.NewDefaultQueue(100),
		alertQueue:    watch.NewDefaultQueue(100),
		alertStorage:  watch.NewDefaultStorage(),
	}

	cacheConf := c.GetData().GetCache()
	if !types.IsNil(cacheConf) {
		d.cacher = newCache(cacheConf)
		d.alertStorage = watch.NewCacheStorage(d.cacher)
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close alert storage", d.alertStorage.Close())
			log.Debugw("close cache", d.cacher.Close())
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

// newCache new cache
func newCache(c *houyiconf.Data_Cache) cache.ICacher {
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
