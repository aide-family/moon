package notify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/wire"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/Alioth/app/merak/internal/conf"
	"github.com/aide-family/Alioth/app/merak/internal/data/repository"
	"github.com/aide-family/Alioth/app/merak/pkg/bo"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
	pkgworker "github.com/aide-family/Alioth/pkg/worker"
)

const (
	notifyLockTTL = 60 * time.Second
	renewEvery    = 20 * time.Second
	pollEvery     = 2 * time.Second
	roundTimeout  = 30 * time.Second
	backoffMax    = 30 * time.Second
	pollSpec      = "*/2 * * * * *"
)

var ProviderSet = wire.NewSet(NewWorker)

// Worker 领取任务、持有投递锁并回写。HTTP 健康检查不经过这里。
type Worker struct {
	gateway repository.NotifyGateway
	locks   repository.Locker
	timeout time.Duration
	log     *slog.Logger
	cancel  context.CancelFunc
	done    chan struct{}
	gate    *pkgworker.Gate
}

func NewWorker(gateway repository.NotifyGateway, locks repository.Locker, notify *conf.Notify, logger *slog.Logger) *Worker {
	timeout := 30 * time.Second
	if notify != nil && notify.GetTimeout() != nil {
		timeout = notify.GetTimeout().AsDuration()
	}
	return &Worker{gateway: gateway, locks: locks, timeout: timeout, log: logger, gate: pkgworker.NewGate(pollEvery, backoffMax)}
}

func (w *Worker) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.done = make(chan struct{})
	sched := cron.New(cron.WithSeconds())
	if _, err := sched.AddFunc(pollSpec, func() { w.round(ctx) }); err != nil {
		cancel()
		return merr.Server("notify cron failed")
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
		w.log.Error("notify worker stop timed out")
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
		w.log.Error("notify round timed out")
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
			w.log.Error(fmt.Sprintf("notify redis: %v", err))
			return true
		}
		return errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	tasks, err := w.gateway.Claim(ctx)
	if err != nil {
		if ctx.Err() == nil {
			w.log.Error(fmt.Sprintf("claim: %v", err))
			return true
		}
		return errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			return errors.Is(err, context.DeadlineExceeded)
		}
		if err := w.deliverOne(ctx, task); err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return true
			}
			w.log.Error(fmt.Sprintf("deliver %s: %v", task.UID, err))
		}
	}
	return false
}

func (w *Worker) deliverOne(ctx context.Context, task bo.TaskSnapshot) error {
	hold, err := w.locks.Acquire(ctx, "alioth:lock:notify:"+task.UID, notifyLockTTL)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	return w.finish(ctx, hold, task)
}

func (w *Worker) finish(ctx context.Context, hold repository.Hold, task bo.TaskSnapshot) error {
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
				if err := hold.Renew(ctx, notifyLockTTL); err != nil {
					renewFailed <- err
					cancel()
					return
				}
				timer.Reset(renewEvery)
			}
		}
	}()
	if err := w.gateway.MarkSending(ctx, task.UID); err != nil {
		return err
	}
	deliverCtx, deliverCancel := context.WithTimeout(ctx, w.timeout)
	defer deliverCancel()
	ok, detail := Deliver(deliverCtx, task)
	select {
	case err := <-renewFailed:
		w.log.Error(fmt.Sprintf("renew %s: %v", task.UID, err))
		if ok {
			return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: false, Detail: "lock renew failed"})
		}
		return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: false, Detail: detail})
	default:
	}
	if ctx.Err() != nil && ok {
		detail := "lock renew failed"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			detail = "round timed out"
		}
		return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: false, Detail: detail})
	}
	return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: ok, Detail: detail})
}
