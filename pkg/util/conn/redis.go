package conn

import (
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/durationpb"
)

// RedisConfig redis配置
type RedisConfig interface {
	GetNetwork() string
	GetAddr() string
	GetPassword() string
	GetDb() uint32
	GetWriteTimeout() *durationpb.Duration
	GetReadTimeout() *durationpb.Duration
	GetDialTimeout() *durationpb.Duration
}

// NewRedisClient 获取redis客户端
func NewRedisClient(cfg RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddr(),
		Password:     cfg.GetPassword(),
		DB:           int(cfg.GetDb()),
		WriteTimeout: cfg.GetWriteTimeout().AsDuration(),
		ReadTimeout:  cfg.GetReadTimeout().AsDuration(),
		DialTimeout:  cfg.GetDialTimeout().AsDuration(),
		Network:      cfg.GetNetwork(),
	})
}
