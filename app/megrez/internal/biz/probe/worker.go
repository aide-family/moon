// Package probe 执行探测规则：领取「规则 × 目标组 × 目标」，按目标加锁后发
// HTTP 请求并回写结果。判定在玉衡，执行器不解析级别条件。
package probe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/app/megrez/internal/data/repository"
	"github.com/aide-family/moon/app/megrez/pkg/bo"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
	pkgworker "github.com/aide-family/moon/pkg/worker"
)

var ProviderSet = wire.NewSet(NewWorker)

const (
	// pollEvery 是探测轮询周期。
	pollEvery = 5 * time.Second
	// evalLockTTL 是单目标锁的租期。
	probeLockTTL = 30 * time.Second
	// roundTimeout 是单轮上限：没跑完的目标留到下一轮（槽位不同，不会重复）。
	roundTimeout = 30 * time.Second
	// backoffMax 是连续失败后的最大退避。
	backoffMax = 30 * time.Second
	// heartbeatEvery 是实例状态上报间隔。
	heartbeatEvery = 5 * time.Second
	// pollSpec 是 cron 表达式，与 pollEvery 对应。
	pollSpec = "*/5 * * * * *"
)

// Worker 是 megrez 里的探测执行器，与评测 worker 并列。
type Worker struct {
	gateway repository.ProbeGateway
	locks   repository.Locker
	log     *slog.Logger
	now     func() time.Time
	gate    *pkgworker.Gate

	mu            sync.RWMutex
	reportMu      sync.Mutex
	instanceUID   string
	hostname      string
	busy          bool
	taskName      string
	taskStartedAt time.Time

	cancel context.CancelFunc
	done   chan struct{}
}

func NewWorker(gateway repository.ProbeGateway, locks repository.Locker, logger *slog.Logger) *Worker {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	instanceUID := fmt.Sprintf("%s-%d-probe-%d", hostname, os.Getpid(), time.Now().UnixNano())
	return &Worker{
		gateway: gateway, locks: locks, log: logger, now: time.Now,
		gate: pkgworker.NewGate(pollEvery, backoffMax), instanceUID: instanceUID, hostname: hostname,
	}
}

func (w *Worker) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.done = make(chan struct{})
	sched := cron.New(cron.WithSeconds())
	if _, err := sched.AddFunc(pollSpec, func() { w.round(ctx) }); err != nil {
		cancel()
		return merr.Server("probe cron failed")
	}
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		w.heartbeat(ctx)
	}()
	go func() {
		defer close(w.done)
		defer func() { <-heartbeatDone }()
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
		w.log.Error("probe worker stop timed out")
	}
	return nil
}

func (w *Worker) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(heartbeatEvery)
	defer ticker.Stop()
	for {
		w.reportStatus(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// reportStatus 用同一个 MEGREZ 身份上报，任务名区分探测。
func (w *Worker) reportStatus(parent context.Context) {
	w.reportMu.Lock()
	defer w.reportMu.Unlock()
	w.mu.RLock()
	status := bo.WorkerStatus{
		InstanceUID: w.instanceUID, Hostname: w.hostname, Busy: w.busy,
		TaskName: w.taskName, TaskStartedAt: w.taskStartedAt,
		PollIntervalSeconds: int32(pollEvery.Seconds()),
	}
	w.mu.RUnlock()
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	if err := w.gateway.ReportStatus(ctx, status); err != nil && parent.Err() == nil {
		w.log.Error(fmt.Sprintf("report probe status: %v", err))
	}
}

func (w *Worker) setBusy(taskName string, startedAt time.Time) {
	w.mu.Lock()
	w.busy = taskName != ""
	w.taskName = taskName
	w.taskStartedAt = startedAt
	w.mu.Unlock()
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
		w.log.Error("probe round timed out")
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
			w.log.Error(fmt.Sprintf("probe redis: %v", err))
			return true
		}
		return errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	tasks, err := w.gateway.ClaimProbeTasks(ctx, "CLUSTER")
	if err != nil {
		if ctx.Err() == nil {
			w.log.Error(fmt.Sprintf("claim probe tasks: %v", err))
			return true
		}
		return errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			return errors.Is(err, context.DeadlineExceeded)
		}
		if err := w.runTask(ctx, task); err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return true
			}
			w.log.Error(fmt.Sprintf("probe task %s/%s: %v", task.Rule.UID, task.Group.UID, err))
		}
	}
	return false
}

// runTask 探一个目标组在本轮的全部分目标，然后一次性回写。
func (w *Worker) runTask(ctx context.Context, task bo.ProbeTask) error {
	w.setBusy(taskName(task), w.now().UTC())
	defer w.setBusy("", time.Time{})
	results := make([]bo.ProbeResult, 0, len(task.Targets))
	for _, target := range task.Targets {
		hold, err := w.locks.Acquire(ctx, lockKey(target.UID, task.Slot), probeLockTTL)
		if err != nil {
			if errors.Is(err, lock.ErrNotAcquired) {
				// 别的副本正在探这个目标，本轮跳过。
				continue
			}
			return err
		}
		result := execute(ctx, task.Group, target, task.Rule.ProbeMetric)
		result.Executor = w.hostname
		results = append(results, result)
		_ = hold.Release(context.Background())
	}
	if len(results) == 0 {
		return nil
	}
	return w.gateway.ReportProbeResults(ctx, bo.ProbeReport{
		RuleUID: task.Rule.UID, Slot: task.Slot, Results: results,
	})
}

func taskName(task bo.ProbeTask) string {
	name := task.Rule.Name
	if name == "" {
		name = task.Rule.UID
	}
	return "probe:" + name + "@" + task.Group.Name
}

// lockKey 用目标与槽位做锁，多个副本同时领取时同一目标只探一次。
func lockKey(targetUID string, slot time.Time) string {
	return fmt.Sprintf("alioth:lock:probe:%s:%d", targetUID, slot.Unix())
}
