// Package probesla 把探测结果按自然日聚合并出具 SLA 报表。
package probesla

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// RollupLockKey 是日聚合任务的锁，与其他调度锁一样按副本互斥。
	RollupLockKey = "alioth:lock:schedule:probe-sla"
	// rollupLockTTL 覆盖一次 SQL 聚合的时间；拿到锁后不续期。
	rollupLockTTL = 30 * time.Second
	// rollupWindow 是每次重算的回溯窗口，覆盖原始探测结果的保留期。
	rollupWindow = 7 * 24 * time.Hour
	// MaxRangeDays 是 SLA 报表允许的最大跨度。
	MaxRangeDays = 30
)

// Gateway 出具探测 SLA 报表并维护日聚合。
type Gateway struct {
	sla    repository.ProbeSlaRepo
	access access.Checker
	locks  *lock.Client
	log    *slog.Logger
}

func NewGateway(sla repository.ProbeSlaRepo, checker access.Checker, locks *lock.Client, logger *slog.Logger) *Gateway {
	return &Gateway{sla: sla, access: checker, locks: locks, log: logger}
}

// Report 校验日期范围后返回按目标组汇总的 SLA 报表。
func (gateway *Gateway) Report(ctx context.Context, groupUIDs []string, from, to time.Time) (bo.ProbeSlaReport, error) {
	ns, err := gateway.access.Allow(ctx, perm.ProbeSlaRead)
	if err != nil {
		return bo.ProbeSlaReport{}, err
	}
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return bo.ProbeSlaReport{}, merr.Invalid("invalid time range")
	}
	if to.Sub(from) > MaxRangeDays*24*time.Hour {
		return bo.ProbeSlaReport{}, merr.Invalid("time range exceeds 30 days")
	}
	report, err := gateway.sla.Report(ctx, ns, groupUIDs, from.UTC(), to.UTC())
	return report, gateway.access.Wrap(err)
}

// Rollup 在锁内重算最近 7 天的日聚合。拿不到锁说明其他副本在做，直接跳过。
func (gateway *Gateway) Rollup(ctx context.Context) error {
	if gateway.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := gateway.locks.Acquire(ctx, RollupLockKey, rollupLockTTL)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	return gateway.sla.Rollup(ctx, time.Now().UTC().Add(-rollupWindow))
}
