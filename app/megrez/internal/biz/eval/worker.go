package eval

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/app/megrez/internal/data/repository"
	"github.com/aide-family/moon/app/megrez/pkg/bo"
	"github.com/aide-family/moon/pkg/alerttmpl"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
	pkgworker "github.com/aide-family/moon/pkg/worker"
)

const (
	evalLockTTL    = 30 * time.Second
	renewEvery     = 10 * time.Second
	pollEvery      = 2 * time.Second
	heartbeatEvery = 5 * time.Second
	roundTimeout   = 30 * time.Second
	backoffMax     = 30 * time.Second
	pollSpec       = "*/2 * * * * *"
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

	mu            sync.RWMutex
	reportMu      sync.Mutex
	instanceUID   string
	hostname      string
	busy          bool
	taskUID       string
	taskName      string
	taskStartedAt time.Time
	// tasks 是最近一批领取的「规则 × 数据源」组合，batchStartedAt 是领取时间。
	// 空轮次不覆盖它，这样页面能看到上一批的回写和离开状态。
	tasks          []bo.WorkerTask
	batchStartedAt time.Time
}

func NewWorker(gateway repository.EvalGateway, locks repository.Locker, query repository.RangeQuery, logger *slog.Logger) *Worker {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	instanceUID := fmt.Sprintf("%s-%d-%d", hostname, os.Getpid(), time.Now().UnixNano())
	return &Worker{
		gateway: gateway, locks: locks, query: query, log: logger, now: time.Now,
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
		return merr.Server("eval cron failed")
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
		w.log.Error("eval worker stop timed out")
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

func (w *Worker) setStatus(ctx context.Context, busy bool, taskUID, taskName string, startedAt time.Time) {
	w.mu.Lock()
	w.busy = busy
	w.taskUID = taskUID
	w.taskName = taskName
	w.taskStartedAt = startedAt
	w.mu.Unlock()
	w.reportStatus(ctx)
}

func (w *Worker) reportStatus(parent context.Context) {
	w.reportMu.Lock()
	defer w.reportMu.Unlock()
	nextPollAt := w.nextPollAt(w.now())
	w.mu.RLock()
	tasks := make([]bo.WorkerTask, len(w.tasks))
	copy(tasks, w.tasks)
	status := bo.WorkerStatus{
		InstanceUID: w.instanceUID, Hostname: w.hostname, Busy: w.busy,
		TaskUID: w.taskUID, TaskName: w.taskName, TaskStartedAt: w.taskStartedAt,
		Tasks: tasks, RoundStartedAt: w.batchStartedAt, NextPollAt: nextPollAt,
		PollIntervalSeconds: int32(pollEvery.Seconds()),
	}
	w.mu.RUnlock()
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	if err := w.gateway.ReportStatus(ctx, status); err != nil && parent.Err() == nil {
		w.log.Error(fmt.Sprintf("report worker status: %v", err))
	}
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
	w.setBatch(ctx, rules)
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
	name := rule.Name
	if name == "" {
		name = rule.UID
	}
	w.markTask(rule, bo.WorkerTaskRunning)
	w.setStatus(ctx, true, rule.UID, name, w.now().UTC())
	defer w.setStatus(context.Background(), false, "", "", time.Time{})
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
		EvalSlot: rule.EvalSlot, Skipped: !numericDriver(rule.Driver),
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
		// 本轮被取消（进程退出或轮次超时）不是数据源的问题，保持原行为不记账。
		if ctx.Err() != nil {
			return err
		}
		// 无效表达式等客户端错误属于规则配置问题，不是数据源不可用：
		// 不回写批次，下一轮继续尝试，用户修正规则后即可恢复。
		if merr.IsInvalid(err) {
			return err
		}
		// 连接失败、超时、上游错误或响应无法解析：回写失败批次，由玉衡累计
		// 数据源可用性；同时让本批次记为已回写，避免每 2 秒重复领取同一条规则。
		report.Failed = true
		report.FailureMessage = err.Error()
		return w.gateway.Report(ctx, report)
	}
	if !report.Skipped {
		var templateQuery TemplateQuery
		if instant, ok := w.query.(repository.InstantQuery); ok {
			templateQuery = func(ctx context.Context, expr string, at time.Time) ([]alerttmpl.Sample, error) {
				queried, err := instant.QueryInstant(ctx, rule, expr, at)
				if err != nil {
					return nil, err
				}
				out := make([]alerttmpl.Sample, 0, len(queried))
				for _, item := range queried {
					if len(item.Points) == 0 {
						continue
					}
					point := item.Points[len(item.Points)-1]
					out = append(out, alerttmpl.Sample{Labels: item.Labels, Value: point.Value})
				}
				return out, nil
			}
		}
		report.Hits = HitsWithQuery(ctx, rule, samples, rule.EvalSlot, templateQuery)
	}
	if err := w.gateway.Report(ctx, report); err != nil {
		w.markTask(rule, bo.WorkerTaskPending)
		return err
	}
	w.markTask(rule, bo.WorkerTaskDone)
	return nil
}

// setBatch 记录本轮领取的组合，并标出相对上一批的新加入和已离开。
// 空轮次不覆盖上一批：真正领到规则的那一轮才有可比对的差异。
func (w *Worker) setBatch(ctx context.Context, rules []bo.RuleSnapshot) {
	if len(rules) == 0 {
		return
	}
	w.mu.Lock()
	previous := make(map[string]bo.WorkerTask, len(w.tasks))
	for _, task := range w.tasks {
		if task.State == bo.WorkerTaskLeft {
			continue
		}
		previous[workerTaskKey(task.RuleUID, task.DatasourceUID)] = task
	}
	tasks := make([]bo.WorkerTask, 0, len(rules)+len(previous))
	for _, rule := range rules {
		key := workerTaskKey(rule.UID, rule.DatasourceUID)
		state := bo.WorkerTaskPending
		if _, ok := previous[key]; !ok {
			state = bo.WorkerTaskNew
		}
		delete(previous, key)
		tasks = append(tasks, bo.WorkerTask{
			State: state, RuleUID: rule.UID, RuleName: rule.Name,
			DatasourceUID: rule.DatasourceUID, DatasourceName: rule.DatasourceName,
			EvalSlot: rule.EvalSlot.UTC(),
		})
	}
	// 已离开的组合保留到下一批领取，排序保证页面上的顺序稳定。
	left := make([]bo.WorkerTask, 0, len(previous))
	for _, task := range previous {
		task.State = bo.WorkerTaskLeft
		left = append(left, task)
	}
	sort.Slice(left, func(i, j int) bool {
		if left[i].RuleUID != left[j].RuleUID {
			return left[i].RuleUID < left[j].RuleUID
		}
		return left[i].DatasourceUID < left[j].DatasourceUID
	})
	w.tasks = append(tasks, left...)
	w.batchStartedAt = w.now().UTC()
	w.mu.Unlock()
	w.reportStatus(ctx)
}

// markTask 更新一个组合的状态。已经离开的组合不参与本次领取，忽略即可。
func (w *Worker) markTask(rule bo.RuleSnapshot, state string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := range w.tasks {
		if w.tasks[i].RuleUID == rule.UID && w.tasks[i].DatasourceUID == rule.DatasourceUID {
			if w.tasks[i].State == bo.WorkerTaskLeft {
				return
			}
			w.tasks[i].State = state
			return
		}
	}
}

// workerTaskKey 用规则和数据源标识一个组合；批次变化不算新的组合。
func workerTaskKey(ruleUID, datasourceUID string) string {
	return ruleUID + "/" + datasourceUID
}

// nextPollAt 返回下一次轮询落在哪个墙钟刻度上。轮询 cron 与墙钟对齐，
// 所以正常情况是下一个 pollEvery 刻度；退避期间顺延到退避结束后的那一拍。
func (w *Worker) nextPollAt(now time.Time) time.Time {
	earliest := now
	if w.gate != nil {
		if wait := w.gate.NextWait(); wait > 0 {
			earliest = now.Add(wait)
		}
	}
	next := earliest.Truncate(pollEvery)
	if !next.After(earliest) {
		next = next.Add(pollEvery)
	}
	return next
}
