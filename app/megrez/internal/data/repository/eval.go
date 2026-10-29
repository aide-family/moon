package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
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
	ReportStatus(ctx context.Context, status bo.WorkerStatus) error
}

// ProbeGateway 领取探测任务并回写结果。与评测共用同一个 alioth 服务与 service key。
type ProbeGateway interface {
	ClaimProbeTasks(ctx context.Context, kind string) ([]bo.ProbeTask, error)
	ReportProbeResults(ctx context.Context, report bo.ProbeReport) error
	ReportStatus(ctx context.Context, status bo.WorkerStatus) error
}

// RangeQuery 向指标数据源查询区间数据。
type RangeQuery interface {
	QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error)
}

// InstantQuery 为 Prometheus 模板中的 query 函数查询瞬时向量。
type InstantQuery interface {
	QueryInstant(ctx context.Context, item bo.RuleSnapshot, expr string, at time.Time) ([]bo.Sample, error)
}
