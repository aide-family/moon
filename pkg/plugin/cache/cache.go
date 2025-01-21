package cache

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const (
	CacherDriverRedis        = "redis"
	CacherDriverRedisCluster = "miniredis"
)

type (
	// ICloser 关闭缓存
	ICloser interface {
		Close() error
		Client() *redis.Client
	}

	// IIntCacher 整数缓存
	IIntCacher interface {
		ICloser
		// IncMax 增加缓存值，如果超过max则返回max
		IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error)
		// DecMin 减少缓存值，如果小于min则返回min
		DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error)
	}

	// ICacher 缓存
	ICacher interface {
		IIntCacher
	}
)

// NewCache new cache
func NewCache(c *conf.Cache) ICacher {
	switch strings.ToLower(c.GetDriver()) {
	case CacherDriverRedis:
		log.Debugw("msg", "redis cache init")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types.IsNil(err) {
			log.Warnw("redis ping error", err)
			panic(err)
		}
		return NewRedisCacher(cli)
	default:
		log.Debugw("msg", "miniredis cache init")
		cli, err := conn.NewMiniRedis()
		if !types.IsNil(err) {
			log.Errorw("miniredis", "init error", "err", err)
			panic(err)
		}
		return NewRedisCacherByMiniRedis(cli)
	}
}
