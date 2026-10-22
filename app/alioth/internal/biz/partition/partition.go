package partition

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// MaintenanceInterval 是分区维护周期。分区按周存在，提前 4 周创建，每小时检查一次足够。
	MaintenanceInterval = time.Hour
	// LockKey 与其他调度锁一样按副本互斥；DDL 用 IF NOT EXISTS，重复执行安全。
	LockKey = "alioth:lock:schedule:partition"
	// lockTTL 覆盖两张表的 CREATE TABLE 时间；拿到锁后不续期。
	lockTTL = 15 * time.Second
)

var ProviderSet = wire.NewSet(NewMaintainer)

// Maintainer 预创建 alert_events 与 message_logs 的周分区。
// 分区创建从写路径移走后，写请求不再执行 DDL；维护任务停摆最多影响未来第 4 周的写入。
type Maintainer struct {
	repo  repository.PartitionRepo
	locks *lock.Client
	log   *slog.Logger
}

func NewMaintainer(repo repository.PartitionRepo, locks *lock.Client, logger *slog.Logger) *Maintainer {
	return &Maintainer{repo: repo, locks: locks, log: logger}
}

// Maintain 在锁内补齐未来 4 周分区。没拿到锁说明另一个副本在做，直接跳过。
func (maintainer *Maintainer) Maintain(ctx context.Context) error {
	if maintainer.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := maintainer.locks.Acquire(ctx, LockKey, lockTTL)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	return maintainer.repo.EnsurePartitions(ctx, time.Now().UTC())
}
