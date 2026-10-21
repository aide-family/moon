package data

import (
	"fmt"
	"log/slog"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
)

func NewLock(c *conf.Data, logger *slog.Logger) (*lock.Client, func(), error) {
	if c == nil || c.Redis == nil || c.Redis.Addr == "" {
		return nil, nil, merr.Server("redis addr is empty")
	}
	rdb := lock.NewRedisClient(c.Redis.Addr, c.Redis.Password)
	cleanup := func() {
		if err := rdb.Close(); err != nil {
			logger.Error(fmt.Sprintf("close redis: %v", err))
		}
	}
	return lock.New(rdb, "alioth"), cleanup, nil
}
