package data

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
)

// partitionRepo 只做分区维护，不参与业务读写。
type partitionRepo struct {
	db  *sql.DB
	log *slog.Logger
}

func NewPartitionRepo(deps *Dependencies, logger *slog.Logger) repository.PartitionRepo {
	return &partitionRepo{db: deps.db, log: logger}
}

// EnsurePartitions 预创建 4 周分区。迁移已建当前周与下一周，
// 这里多建几周是为了让维护任务停摆或发布间隔较长时写路径仍然有分区可用。
func (r *partitionRepo) EnsurePartitions(ctx context.Context, now time.Time) error {
	for _, name := range partitionedTables {
		if err := ensureUpcomingPartitions(ctx, r.db, name, now, partitionLeadWeeks); err != nil {
			r.log.Error("ensure partition " + name + ": " + err.Error())
			return err
		}
	}
	return nil
}
