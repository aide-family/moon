package cacher

import (
	"errors"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/cacher/nutsdbcache"
	"github.com/aide-family/moon/pkg/agent/cacher/rediscache"
	"github.com/nutsdb/nutsdb"
	"github.com/redis/go-redis/v9"
)

type (
	cacheBuild struct {
		redisCli *redis.Client
		nutsDB   *nutsdb.DB
	}

	Option func(c *cacheBuild)
)

var ErrNoCache = errors.New("no cache")

func New(opts ...Option) (agent.Cache, error) {
	c := &cacheBuild{}
	for _, opt := range opts {
		opt(c)
	}
	return c.Builder()
}

// Builder 构建缓存
func (c *cacheBuild) Builder() (agent.Cache, error) {
	if c.redisCli != nil {
		return rediscache.NewRedisCache(c.redisCli), nil
	}
	if c.nutsDB != nil {
		return nutsdbcache.NewNutsDbCache(c.nutsDB, "default"), nil
	}
	return nil, ErrNoCache
}
