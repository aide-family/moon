package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// WorkerRepo persists recent heartbeats from background worker processes.
type WorkerRepo interface {
	Report(ctx context.Context, status bo.WorkerStatus, now time.Time) error
	ListRecent(ctx context.Context, since time.Time) ([]bo.WorkerInstance, error)
	// Get 读取一个实例的最近心跳，含本批「规则 × 数据源」组合快照。
	Get(ctx context.Context, service, instanceUID string) (*bo.WorkerInstance, error)
	CleanupBefore(ctx context.Context, cutoff time.Time) error
}
