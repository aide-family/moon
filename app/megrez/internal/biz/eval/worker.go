package eval

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/wire"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
	pkgworker "github.com/aide-family/Alioth/pkg/worker"
)

const (
	evalLockTTL  = 30 * time.Second
	renewEvery   = 10 * time.Second
	pollEvery    = 2 * time.Second
	roundTimeout = 30 * time.Second
	backoffMax   = 30 * time.Second
	pollSpec     = "*/2 * * * * *"
)

var ProviderSet = wire.NewSet(NewWorker)

// Worker 领取规则、持有评测锁并回写。HTTP 健康检查不经过这里。
type Worker struct {
	gateway repository.EvalGateway
	locks   repository.Locker
	query   repository.RangeQuery
	log     *slog.Logger
	cancel  context.CancelFunc
	done    chan struct{}
	now     func() time.Time
	gate    *pkgworker.Gate
}

func NewWorker(gateway repository.EvalGateway, locks repository.Locker, query repository.RangeQuery, logger *slog.Logger) *Worker {
	return &Worker{gateway: gateway, locks: locks, query: query, log: logger, now: time.Now, gate: pkgworker.NewGate(pollEvery, backoffMax)}
}

func (w *Worker) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.done = make(chan struct{})
	sched := cron.New(cron.WithSeconds())
	if _, err := sched.AddFunc(pollSpec, func() { w.round(ctx) }); err != nil {
		cancel()
		return merr.Server("eval cron failed")
	}
	go func() {
		defer close(w.done)
		w.round(ctx)
		sched.Start()
		<-ctx.Done()
		<-sched.Stop().Done()
	}()
	return nil
}

func (w *Worker) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	if w.done == nil {
		return nil
	}
	select {
	case <-w.done:
	case <-ctx.Done():
		w.log.Error("eval worker stop timed out")
	}
	return nil
}

func (w *Worker) round(ctx context.Context) {
	if ctx.Err() != nil || !w.gate.Begin() {
		return
	}
	roundCtx, cancel := context.WithTimeout(ctx, roundTimeout)
	done := make(chan struct{})
	stretch := false
	go func() {
		defer close(done)
		stretch = w.once(roundCtx)
	}()
	timedOut := false
	select {
	case <-done:
	case <-roundCtx.Done():
		if ctx.Err() != nil {
			cancel()
			w.gate.Note(false)
			return
		}
		timedOut = true
		w.log.Error("eval round timed out")
	}
	cancel()
	extend := timedOut
	if !timedOut {
		extend = stretch
	}
	w.gate.Note(extend)
}

func (w *Worker) once(ctx context.Context) bool {
	if err := w.locks.Available(ctx); err != nil {
		if ctx.Err() == nil {
			w.log.Error(fmt.Sprintf("eval redis: %v", err))
			return true
		}
		return errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	rules, err := w.gateway.Claim(ctx)
	if err != nil {
		if ctx.Err() == nil {
			w.log.Error(fmt.Sprintf("claim: %v", err))
			return true
		}
		return errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	for _, rule := range rules {
		if err := ctx.Err(); err != nil {
			return errors.Is(err, context.DeadlineExceeded)
		}
		if err := w.evalOne(ctx, rule); err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return true
			}
			w.log.Error(fmt.Sprintf("eval %s: %v", rule.UID, err))
		}
	}
	return false
}

func (w *Worker) evalOne(ctx context.Context, rule bo.RuleSnapshot) error {
	hold, err := w.locks.Acquire(ctx, "alioth:lock:eval:"+rule.UID+":"+rule.DatasourceUID, evalLockTTL)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	return w.finish(ctx, hold, rule)
}

func (w *Worker) finish(ctx context.Context, hold repository.Hold, rule bo.RuleSnapshot) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	renewFailed := make(chan error, 1)
	go func() {
		timer := time.NewTimer(renewEvery)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if err := hold.Renew(ctx, evalLockTTL); err != nil {
					renewFailed <- err
					cancel()
					return
				}
				timer.Reset(renewEvery)
			}
		}
	}()
	report := bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: rule.DatasourceUID,
		EvalSlot: rule.EvalSlot, Skipped: !metricsDriver(rule.Driver),
	}
	var samples []bo.Sample
	var err error
	if !report.Skipped {
		start, end, step := window(rule, w.now())
		samples, err = w.query.QueryRange(ctx, rule, start, end, step)
	}
	select {
	case renewErr := <-renewFailed:
		return renewErr
	default:
	}
	if renewErr := hold.Renew(ctx, evalLockTTL); renewErr != nil {
		w.log.Error(fmt.Sprintf("renew %s: %v", rule.UID, renewErr))
		return merr.Server("lock failed")
	}
	if err != nil {
		return err
	}
	if !report.Skipped {
		report.Hits = Hits(rule, samples)
	}
	if err := w.gateway.Report(ctx, report); err != nil {
		return err
	}
	return nil
}
