package repository

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
)

// Hold 是一把已拿到的评测锁。
type Hold interface {
	Renew(ctx context.Context, ttl time.Duration) error
	Release(ctx context.Context) error
}

// Locker 只通过分布式锁客户端获取评测锁。
type Locker interface {
	Available(ctx context.Context) error
	Acquire(ctx context.Context, key string, ttl time.Duration) (Hold, error)
}

// EvalGateway 向玉衡领取规则并回写结果。
type EvalGateway interface {
	Claim(ctx context.Context) ([]bo.RuleSnapshot, error)
	Report(ctx context.Context, report bo.EvalReport) error
}

// RangeQuery 向指标数据源查询区间数据。
type RangeQuery interface {
	QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error)
}
