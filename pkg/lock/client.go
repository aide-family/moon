package lock

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 构造只用于分布式锁的 Redis 客户端。
// 超时和连接池必须显式设置：默认值在 Redis 抖动时会把请求一起拖住，
// 多副本下也更难判断是锁失败还是网络阻塞。
func NewRedisClient(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MaxRetries:   2,
	})
}
