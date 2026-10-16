package notify

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/conf"
	"github.com/aide-family/Alioth/app/merak/internal/data/repository"
	"github.com/aide-family/Alioth/app/merak/pkg/bo"
	"github.com/aide-family/Alioth/pkg/lock"
)

const (
	notifyLockTTL = 60 * time.Second
	renewEvery    = 20 * time.Second
	pollEvery     = 2 * time.Second
)

var ProviderSet = wire.NewSet(NewWorker)

// Worker 领取任务、持有投递锁并回写。HTTP 健康检查不经过这里。
type Worker struct {
	gateway repository.NotifyGateway
	locks   repository.Locker
	timeout time.Duration
	log     *log.Helper
	cancel  context.CancelFunc
	done    chan struct{}
}

func NewWorker(gateway repository.NotifyGateway, locks repository.Locker, notify *conf.Notify, logger log.Logger) *Worker {
	timeout := 30 * time.Second
	if notify != nil && notify.GetTimeout() != nil {
		timeout = notify.GetTimeout().AsDuration()
	}
	return &Worker{gateway: gateway, locks: locks, timeout: timeout, log: log.NewHelper(logger)}
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
		w.log.Errorf("notify redis: %v", err)
		return
	}
	tasks, err := w.gateway.Claim(ctx)
	if err != nil {
		w.log.Errorf("claim: %v", err)
		return
	}
	for _, task := range tasks {
		if err := w.deliverOne(ctx, task); err != nil {
			w.log.Errorf("deliver %s: %v", task.UID, err)
		}
	}
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
		ticker := time.NewTicker(renewEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := hold.Renew(ctx, notifyLockTTL); err != nil {
					renewFailed <- err
					cancel()
					return
				}
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
		w.log.Errorf("renew %s: %v", task.UID, err)
		if ok {
			return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: false, Detail: "lock renew failed"})
		}
		return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: false, Detail: detail})
	default:
	}
	if ctx.Err() != nil && ok {
		return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: false, Detail: "lock renew failed"})
	}
	return w.gateway.Report(context.Background(), bo.NotifyReport{TaskUID: task.UID, OK: ok, Detail: detail})
}
