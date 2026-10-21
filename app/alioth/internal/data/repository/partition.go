package repository

import (
	"context"
	"time"
)

// PartitionRepo 维护按周分区的事件表与消息日志表。
// 分区创建从写路径移到定时任务：写请求不再执行 DDL，也不会因为当期分区缺失而失败。
type PartitionRepo interface {
	// EnsurePartitions 预创建当前周及未来若干周的分区，可重复执行。
	EnsurePartitions(ctx context.Context, now time.Time) error
}
