package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/pkg/healthcheck"
	"github.com/aide-family/Alioth/pkg/lock"
)

// NewReadinessChecker 汇总控制面的依赖探测：PostgreSQL 与 Redis。
// 只有 alioth 直连业务库，所以 readiness 也只在这里检查数据库。
func NewReadinessChecker(deps *Dependencies, locks *lock.Client) healthcheck.Checker {
	checks := healthcheck.NewChecks(2 * time.Second)
	if deps != nil && deps.db != nil {
		checks.Add("postgres", func(ctx context.Context) error {
			return deps.db.PingContext(ctx)
		})
	}
	if locks != nil {
		checks.Add("redis", locks.Available)
	}
	return checks
}
