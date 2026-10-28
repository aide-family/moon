package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/biz/notification"
	"github.com/aide-family/moon/app/alioth/internal/biz/partition"
	workerbiz "github.com/aide-family/moon/app/alioth/internal/biz/worker"
	"github.com/aide-family/moon/pkg/merr"
)

// Scheduler 按固定周期扫描到期规则。锁由 biz 层获取，扫描结束后释放。
type Scheduler struct {
	coordinator       *alertevent.Coordinator
	center            *notification.Center
	maintainer        *partition.Maintainer
	roster            *workerbiz.Roster
	log               *slog.Logger
	lastCheck         time.Time
	lastWorkerCleanup time.Time
	lastBackfill      time.Time
	lastLogBackfill   time.Time
	cancel            context.CancelFunc
	done              chan struct{}
}

func NewScheduler(
	coordinator *alertevent.Coordinator,
	center *notification.Center,
	maintainer *partition.Maintainer,
	roster *workerbiz.Roster,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{coordinator: coordinator, center: center, maintainer: maintainer, roster: roster, log: logger}
}

func (s *Scheduler) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})
	spec := fmt.Sprintf("*/%d * * * * *", int(alertevent.EvalInterval.Seconds()))
	sched := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))
	if _, err := sched.AddFunc(spec, func() { s.tick(ctx) }); err != nil {
		cancel()
		return merr.Server("schedule cron failed")
	}
	go func() {
		defer close(s.done)
		s.tick(ctx)
		sched.Start()
		<-ctx.Done()
		<-sched.Stop().Done()
	}()
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.done == nil {
		return nil
	}
	select {
	case <-s.done:
	case <-ctx.Done():
		s.log.Error("scheduler stop timed out")
	}
	return nil
}

func (s *Scheduler) tick(ctx context.Context) {
	if err := s.coordinator.Schedule(ctx); err != nil {
		s.log.Error(fmt.Sprintf("schedule eval: %v", err))
	}
	if err := s.center.ReapStale(ctx); err != nil {
		s.log.Error(fmt.Sprintf("reap notify: %v", err))
	}
	s.maintainPartitions(ctx)
	s.backfillHistory(ctx)
	s.backfillMessageLogs(ctx)
	s.cleanupWorkerInstances(ctx)
}

// cleanupWorkerInstances periodically prunes stopped process identities independently of dashboard traffic.
func (s *Scheduler) cleanupWorkerInstances(ctx context.Context) {
	if s.roster == nil || time.Since(s.lastWorkerCleanup) < time.Hour {
		return
	}
	s.lastWorkerCleanup = time.Now()
	if err := s.roster.CleanupStale(ctx); err != nil {
		s.log.Error(fmt.Sprintf("cleanup worker instances: %v", err))
	}
}

// backfillHistory 每小时把主库遗留的已恢复事件写入历史分库分表。
// 首次 tick 会执行，已有部署升级后历史告警会逐步迁入租户库。
func (s *Scheduler) backfillHistory(ctx context.Context) {
	if time.Since(s.lastBackfill) < partition.MaintenanceInterval {
		return
	}
	s.lastBackfill = time.Now()
	if _, err := s.coordinator.BackfillHistory(ctx); err != nil {
		s.log.Error(fmt.Sprintf("backfill alert history: %v", err))
	}
}

// backfillMessageLogs 每小时把主库遗留的消息日志写入租户分库分表。
// 新日志在投递回写后立即归档，这里兜底升级前的数据和归档失败的重试。
func (s *Scheduler) backfillMessageLogs(ctx context.Context) {
	if time.Since(s.lastLogBackfill) < partition.MaintenanceInterval {
		return
	}
	s.lastLogBackfill = time.Now()
	if _, err := s.center.BackfillMessageLogs(ctx); err != nil {
		s.log.Error(fmt.Sprintf("backfill message logs: %v", err))
	}
}

// maintainPartitions 每小时预创建一次周分区。首次 tick 会立即执行，
// 保证新部署在写入前就有当前周与未来 4 周的分区。
func (s *Scheduler) maintainPartitions(ctx context.Context) {
	if s.maintainer == nil || time.Since(s.lastCheck) < partition.MaintenanceInterval {
		return
	}
	s.lastCheck = time.Now()
	if err := s.maintainer.Maintain(ctx); err != nil {
		s.log.Error(fmt.Sprintf("maintain partitions: %v", err))
	}
}
