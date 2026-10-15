package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertevent"
)

// Scheduler 按固定周期扫描到期规则。锁由用例获取，扫描结束后释放。
type Scheduler struct {
	uc     *alertevent.Usecase
	log    *log.Helper
	cancel context.CancelFunc
	done   chan struct{}
}

func NewScheduler(uc *alertevent.Usecase, logger log.Logger) *Scheduler {
	return &Scheduler{uc: uc, log: log.NewHelper(logger)}
}

func (s *Scheduler) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})
	go func() {
		defer close(s.done)
		s.tick(ctx)
		ticker := time.NewTicker(alertevent.EvalInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tick(ctx)
			}
		}
	}()
	return nil
}

func (s *Scheduler) Stop(context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.done != nil {
		<-s.done
	}
	return nil
}

func (s *Scheduler) tick(ctx context.Context) {
	if err := s.uc.Schedule(ctx); err != nil {
		s.log.Errorf("schedule eval: %v", err)
	}
}
