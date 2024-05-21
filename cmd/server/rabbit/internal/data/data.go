package data

import (
	"context"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/conn/cacher/nutsdbcacher"
	"github.com/aide-cloud/moon/pkg/conn/cacher/rediscacher"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	cacher   conn.Cache
	enforcer *casbin.SyncedEnforcer
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
		if err := cli.Ping(context.Background()).Err(); err != nil {
			log.Warnw("redis ping error", err)
		}
		return rediscacher.NewRedisCacher(cli)
	}

	if !types.IsNil(c.GetNutsDB()) {
		log.Debugw("cache init", "nutsdb")
		cli, err := nutsdbcacher.NewNutsDbCacher(c.GetNutsDB())
		if err != nil {
			log.Warnw("nutsdb init error", err)
		}
		return cli
	}
	return nil
}