package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/pkg/healthcheck"
	"github.com/aide-family/Alioth/pkg/lock"
)

// NewReadiness 汇总控制面的依赖探测：PostgreSQL 与 Redis。
// 只有 alioth 直连业务库，所以 readiness 也只在这里检查数据库。
func NewReadiness(d *Data, locks *lock.Client) healthcheck.Checker {
	checks := healthcheck.New(2 * time.Second)
	if d != nil && d.db != nil {
		checks.Add("postgres", func(ctx context.Context) error {
			return d.db.PingContext(ctx)
		})
	}
	if locks != nil {
		checks.Add("redis", locks.Available)
	}
	return checks
}
