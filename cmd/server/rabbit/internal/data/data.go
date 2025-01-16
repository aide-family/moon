package data

import (
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/watch"

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
func NewData(c *rabbitconf.Bootstrap) (*Data, func(), error) {
	d := &Data{
		watchStorage: watch.NewDefaultStorage(),
		watchQueue:   watch.NewDefaultQueue(watch.QueueMaxSize),
	}
	d.cacher = cache.NewCache(c.GetCache())
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
