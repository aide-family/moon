package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/merak/pkg/bo"
)

// NotifyGateway 调用玉衡领取和回写。
type NotifyGateway interface {
	Claim(ctx context.Context) ([]bo.TaskSnapshot, error)
	MarkSending(ctx context.Context, taskUID string) error
	Report(ctx context.Context, report bo.NotifyReport) error
	ReportStatus(ctx context.Context, status bo.WorkerStatus) error
}

// Locker 管理投递锁。
type Locker interface {
	Available(ctx context.Context) error
	Acquire(ctx context.Context, key string, ttl time.Duration) (Hold, error)
}

// Hold 是一把已持有的锁。
type Hold interface {
	Renew(ctx context.Context, ttl time.Duration) error
	Release(ctx context.Context) error
}
