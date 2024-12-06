package data

import (
	"context"
	"strings"

	"github.com/aide-family/moon/cmd/server/jadeTree/internal/jadetreeconf"
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

	watchStorage watch.Storage
	watchQueue   watch.Queue
}

var closeFuncList []func()

// NewData .
func NewData(c *jadetreeconf.Bootstrap) (*Data, func(), error) {
	d := &Data{
		watchStorage: watch.NewDefaultStorage(),
		watchQueue:   watch.NewDefaultQueue(1000),
	}
	d.cacher = newCache(c.GetCache())
	closeFuncList = append(closeFuncList, func() {
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
		panic("cache is nil")
	}
	return d.cacher
}

// GetWatcherStorage 获取 watcher 存储
func (d *Data) GetWatcherStorage() watch.Storage {
	return d.watchStorage
}

// GetWatcherQueue 获取 watcher 队列
func (d *Data) GetWatcherQueue() watch.Queue {
	return d.watchQueue
}

// newCache new cache
func newCache(c *conf.Cache) cache.ICacher {
	switch strings.ToLower(c.GetDriver()) {
	case "redis":
		log.Debugw("cache init", "redis")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types.IsNil(err) {
			log.Warnw("redis ping error", err)
		}
		return cache.NewRedisCacher(cli)
	case "nutsdb":
		log.Debugw("cache init", "nutsdb")
		cli, err := conn.NewNutsDB(c.GetNutsDB())
		if !types.IsNil(err) {
			log.Warnw("nutsdb init error", err)
		}
		return cache.NewNutsDbCacher(cli, c.GetNutsDB().GetBucket())
	default:
		log.Debugw("cache init", "free")
		size := int(c.GetFree().GetSize())
		return cache.NewFreeCache(freecache.NewCache(types.Ternary(size > 0, size, 10*1024*1024)))
	}
}
