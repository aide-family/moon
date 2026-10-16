package eval

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	evalLockTTL = 30 * time.Second
	renewEvery  = 10 * time.Second
	pollEvery   = 2 * time.Second
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
}

func NewWorker(gateway repository.EvalGateway, locks repository.Locker, query repository.RangeQuery, logger *slog.Logger) *Worker {
	return &Worker{gateway: gateway, locks: locks, query: query, log: logger, now: time.Now}
}

func (w *Worker) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.done = make(chan struct{})
	go func() {
		defer close(w.done)
		ticker := time.NewTicker(pollEvery)
		defer ticker.Stop()
		w.once(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.once(ctx)
			}
		}
	}()
	return nil
}

func (w *Worker) Stop(context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	if w.done != nil {
		<-w.done
	}
	return nil
}

func (w *Worker) once(ctx context.Context) {
	if err := w.locks.Available(ctx); err != nil {
		w.log.Error(fmt.Sprintf("eval redis: %v", err))
		return
	}
	rules, err := w.gateway.Claim(ctx)
	if err != nil {
		w.log.Error(fmt.Sprintf("claim: %v", err))
		return
	}
	for _, rule := range rules {
		if err := w.evalOne(ctx, rule); err != nil {
			w.log.Error(fmt.Sprintf("eval %s: %v", rule.UID, err))
		}
	}
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
		ticker := time.NewTicker(renewEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := hold.Renew(ctx, evalLockTTL); err != nil {
					renewFailed <- err
					cancel()
					return
				}
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
