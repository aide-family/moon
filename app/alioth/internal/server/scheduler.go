package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/notification"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/partition"
	"github.com/aide-family/Alioth/pkg/merr"
)

// Scheduler 按固定周期扫描到期规则。锁由用例获取，扫描结束后释放。
type Scheduler struct {
	uc         *alertevent.Usecase
	notify     *notification.Usecase
	partitions *partition.Usecase
	log        *slog.Logger
	lastCheck  time.Time
	cancel     context.CancelFunc
	done       chan struct{}
}

func NewScheduler(uc *alertevent.Usecase, notify *notification.Usecase, partitions *partition.Usecase, logger *slog.Logger) *Scheduler {
	return &Scheduler{uc: uc, notify: notify, partitions: partitions, log: logger}
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
	if err := s.uc.Schedule(ctx); err != nil {
		s.log.Error(fmt.Sprintf("schedule eval: %v", err))
	}
	if err := s.notify.ReapStale(ctx); err != nil {
		s.log.Error(fmt.Sprintf("reap notify: %v", err))
	}
	s.maintainPartitions(ctx)
}

// maintainPartitions 每小时预创建一次周分区。首次 tick 会立即执行，
// 保证新部署在写入前就有当前周与未来 4 周的分区。
func (s *Scheduler) maintainPartitions(ctx context.Context) {
	if s.partitions == nil || time.Since(s.lastCheck) < partition.MaintenanceInterval {
		return
	}
	s.lastCheck = time.Now()
	if err := s.partitions.Maintain(ctx); err != nil {
		s.log.Error(fmt.Sprintf("maintain partitions: %v", err))
	}
}
