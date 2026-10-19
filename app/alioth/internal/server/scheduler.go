package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/Alioth/pkg/merr"
)

// Scheduler 按固定周期扫描到期规则。锁由用例获取，扫描结束后释放。
type Scheduler struct {
	uc     *alertevent.Usecase
	log    *slog.Logger
	cancel context.CancelFunc
	done   chan struct{}
}

func NewScheduler(uc *alertevent.Usecase, logger *slog.Logger) *Scheduler {
	return &Scheduler{uc: uc, log: logger}
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
}
