package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
)

func NewLock(c *conf.Data, logger log.Logger) (*lock.Client, func(), error) {
	if c == nil || c.Redis == nil || c.Redis.Addr == "" {
		return nil, nil, merr.Server("redis addr is empty")
	}
	rdb := redis.NewClient(&redis.Options{Addr: c.Redis.Addr, Password: c.Redis.Password})
	cleanup := func() {
		if err := rdb.Close(); err != nil {
			log.NewHelper(logger).Errorf("close redis: %v", err)
		}
	}
	return lock.New(rdb, "alioth"), cleanup, nil
}
