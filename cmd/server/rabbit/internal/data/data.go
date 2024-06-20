package data

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/conn/cacher/nutsdbcacher"
	"github.com/aide-family/moon/pkg/util/conn/cacher/rediscacher"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	cacher conn.Cache
}

var closeFuncList []func()

// NewData .
func NewData(c *rabbitconf.Bootstrap) (*Data, func(), error) {
	d := &Data{}
	cacheConf := c.GetData().GetCache()
	if !types.IsNil(cacheConf) {
		cacheInstance := newCache(cacheConf)
		if types.IsNil(cacheInstance) {
			return nil, func() {}, merr.ErrorNotification("缓存实例化失败")
		}
		d.cacher = cacheInstance
		closeFuncList = append(closeFuncList, func() {
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
func (d *Data) GetCacher() conn.Cache {
	if types.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// newCache new cache
func newCache(c *rabbitconf.Data_Cache) conn.Cache {
	if types.IsNil(c) {
		return nil
	}

	if !types.IsNil(c.GetRedis()) {
		log.Debugw("cache init", "redis")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types.IsNil(err) {
			log.Warnw("redis ping error", err)
		}
		return rediscacher.NewRedisCacher(cli)
	}

	if !types.IsNil(c.GetNutsDB()) {
		log.Debugw("cache init", "nutsdb")
		cli, err := nutsdbcacher.NewNutsDbCacher(c.GetNutsDB())
		if !types.IsNil(err) {
			log.Warnw("nutsdb init error", err)
		}
		return cli
	}
	return nil
}
