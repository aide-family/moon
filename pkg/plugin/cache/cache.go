package cache

import (
	"context"
	"encoding"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/coocood/freecache"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	CacherDriverRedis        = "redis"
	CacherDriverRedisCluster = "miniredis"
)

type (
	// ICloser 关闭缓存
	ICloser interface {
		Close() error
	}

	// ISimpleCacher 简单缓存
	ISimpleCacher interface {
		ICloser
		// Delete 删除缓存
		Delete(ctx context.Context, key string) error
		// Exist 判断缓存是否存在
		Exist(ctx context.Context, key string) (bool, error)
		// Get 获取缓存
		Get(ctx context.Context, key string) (string, error)
		// Set 设置缓存
		Set(ctx context.Context, key string, value string, expiration time.Duration) error
	}

	// IIntCacher 整数缓存
	IIntCacher interface {
		ICloser
		// Inc 增加缓存值
		Inc(ctx context.Context, key string, expiration time.Duration) (int64, error)
		// Dec 减少缓存值
		Dec(ctx context.Context, key string, expiration time.Duration) (int64, error)
		// IncMax 增加缓存值，如果超过max则返回max
		IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error)
		// DecMin 减少缓存值，如果小于min则返回min
		DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error)
		// GetInt64 获取缓存
		GetInt64(ctx context.Context, key string) (int64, error)
		// SetInt64 设置缓存
		SetInt64(ctx context.Context, key string, value int64, expiration time.Duration) error
	}

	// IBoolCacher 布尔缓存
	IBoolCacher interface {
		ICloser
		GetBool(ctx context.Context, key string) (bool, error)
		SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error
	}

	// IFloatCacher 浮点数缓存
	IFloatCacher interface {
		ICloser
		GetFloat64(ctx context.Context, key string) (float64, error)
		SetFloat64(ctx context.Context, key string, value float64, expiration time.Duration) error
	}

	// IObjectSchema 缓存对象
	IObjectSchema interface {
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	}

	// IObjectCacher 对象缓存
	IObjectCacher interface {
		ICloser
		// GetObject 获取缓存
		GetObject(ctx context.Context, key string, obj IObjectSchema) error
		// SetObject 设置缓存
		SetObject(ctx context.Context, key string, obj IObjectSchema, expiration time.Duration) error
	}

	// IKeyPrefixCacher 带前缀的缓存
	IKeyPrefixCacher interface {
		ICloser
		// Keys 获取带前缀的缓存key
		Keys(ctx context.Context, prefix string) ([]string, error)
		// DelKeys 删除带前缀的缓存key
		DelKeys(ctx context.Context, prefix string) error
	}

	// ICacher 缓存
	ICacher interface {
		ISimpleCacher
		IIntCacher
		IFloatCacher
		IObjectCacher
		IBoolCacher
		IKeyPrefixCacher
		// SetNX 设置缓存，如果key存在则返回false
		SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error)
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
	case CacherDriverRedisCluster:
		log.Debugw("msg", "miniredis cache init")
		cli, err := conn.NewMiniRedis()
		if !types.IsNil(err) {
			log.Errorw("miniredis", "init error", "err", err)
			panic(err)
		}
		return NewRedisCacherByMiniRedis(cli)
	default:
		log.Debugw("msg", "free cache init")
		size := int(c.GetFree().GetSize())
		return NewFreeCache(freecache.NewCache(types.Ternary(size > 0, size, 10*1024*1024)))
	}
}
