package worker

import (
	"context"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	recentWindow    = time.Minute
	onlineWindow    = 15 * time.Second
	recordRetention = 24 * time.Hour
	// maxWorkerTasks 是一次心跳里接受的「规则 × 数据源」组合上限。
	// 超过上限时保留前面的组合：丢几条展示数据也比整条心跳被拒、页面变离线好。
	maxWorkerTasks = 256
	// workerFieldLimit 是实例、规则和数据源标识的长度上限。
	workerFieldLimit = 160
	// workerNameLimit 是规则名、数据源名和任务名的显示长度上限。
	workerNameLimit = 240
)

var ProviderSet = wire.NewSet(NewRoster)

// Roster 接收工作实例心跳，并给出管理员视角的在线实例列表。
type Roster struct {
	repo   repository.WorkerRepo
	access access.Checker
	now    func() time.Time
}

func NewRoster(repo repository.WorkerRepo, checker access.Checker) *Roster {
	return &Roster{repo: repo, access: checker, now: time.Now}
}

// Report stores one process heartbeat. The service name is fixed by the caller's internal gRPC endpoint.
func (roster *Roster) Report(ctx context.Context, service string, status bo.WorkerStatus) error {
	service = strings.ToUpper(strings.TrimSpace(service))
	if service != "MEGREZ" && service != "MERAK" {
		return merr.Client("invalid worker service")
	}
	status.Service = service
	status.InstanceUID = strings.TrimSpace(status.InstanceUID)
	status.Hostname = strings.TrimSpace(status.Hostname)
	status.TaskUID = strings.TrimSpace(status.TaskUID)
	status.TaskName = strings.TrimSpace(status.TaskName)
	if status.InstanceUID == "" || len(status.InstanceUID) > workerFieldLimit {
		return merr.Client("invalid worker instance")
	}
	if !status.Busy {
		status.TaskUID = ""
		status.TaskName = ""
		status.TaskStartedAt = nil
	} else if status.TaskUID == "" || len(status.TaskUID) > workerFieldLimit || status.TaskStartedAt == nil || status.TaskStartedAt.IsZero() {
		return merr.Client("active task details are required")
	}
	status.TaskName = truncateRunes(status.TaskName, workerNameLimit)
	status.Hostname = truncateRunes(status.Hostname, workerFieldLimit)
	// 空闲时也要保留本轮组合：刚回写完的那一批状态（已回写、已离开）正是页面要看的内容。
	tasks, err := normalizeWorkerTasks(status.Tasks)
	if err != nil {
		return err
	}
	status.Tasks = tasks
	status.RoundStartedAt = utcTime(status.RoundStartedAt)
	status.NextPollAt = utcTime(status.NextPollAt)
	return roster.repo.Report(ctx, status, roster.now().UTC())
}

// CleanupStale removes worker records that have exceeded the dashboard retention window.
func (roster *Roster) CleanupStale(ctx context.Context) error {
	return roster.repo.CleanupBefore(ctx, roster.now().UTC().Add(-recordRetention))
}

func (roster *Roster) List(ctx context.Context) ([]bo.WorkerInstance, int, int, int, time.Time, error) {
	if _, err := roster.access.RequireSystem(ctx, perm.SystemUserRead); err != nil {
		return nil, 0, 0, 0, time.Time{}, err
	}
	now := roster.now().UTC()
	items, err := roster.repo.ListRecent(ctx, now.Add(-recentWindow))
	if err != nil {
		return nil, 0, 0, 0, time.Time{}, err
	}
	var online, busy int
	for i := range items {
		items[i].Online = now.Sub(items[i].LastSeenAt) <= onlineWindow
		if items[i].Online {
			online++
			if items[i].Busy {
				busy++
			}
		}
	}
	return items, online, busy, online - busy, now, nil
}

// Get 读取一个实例的本批组合快照，并返回查询时刻。
func (roster *Roster) Get(ctx context.Context, service, instanceUID string) (bo.WorkerInstance, time.Time, error) {
	if _, err := roster.access.RequireSystem(ctx, perm.SystemUserRead); err != nil {
		return bo.WorkerInstance{}, time.Time{}, err
	}
	service = strings.ToUpper(strings.TrimSpace(service))
	if service != "MEGREZ" && service != "MERAK" {
		return bo.WorkerInstance{}, time.Time{}, merr.Client("invalid worker service")
	}
	instanceUID = strings.TrimSpace(instanceUID)
	if instanceUID == "" || len(instanceUID) > workerFieldLimit {
		return bo.WorkerInstance{}, time.Time{}, merr.Client("invalid worker instance")
	}
	item, err := roster.repo.Get(ctx, service, instanceUID)
	if err != nil {
		return bo.WorkerInstance{}, time.Time{}, err
	}
	now := roster.now().UTC()
	item.Online = now.Sub(item.LastSeenAt) <= onlineWindow
	return *item, now, nil
}

// normalizeWorkerTasks 校验并归一化心跳里的组合快照：状态必须已知，
// 规则和数据源不能为空，同一组合只保留一条，批次时间统一成 UTC。
func normalizeWorkerTasks(tasks []bo.WorkerTask) ([]bo.WorkerTask, error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	if len(tasks) > maxWorkerTasks {
		tasks = tasks[:maxWorkerTasks]
	}
	normalized := make([]bo.WorkerTask, 0, len(tasks))
	seen := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		task.State = strings.ToUpper(strings.TrimSpace(task.State))
		if !bo.ValidWorkerTaskState(task.State) {
			return nil, merr.Client("invalid worker task state")
		}
		task.RuleUID = strings.TrimSpace(task.RuleUID)
		task.DatasourceUID = strings.TrimSpace(task.DatasourceUID)
		if task.RuleUID == "" || task.DatasourceUID == "" ||
			len(task.RuleUID) > workerFieldLimit || len(task.DatasourceUID) > workerFieldLimit ||
			task.EvalSlot.IsZero() {
			return nil, merr.Client("invalid worker task")
		}
		key := task.RuleUID + "/" + task.DatasourceUID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		task.RuleName = truncateRunes(task.RuleName, workerNameLimit)
		task.DatasourceName = truncateRunes(task.DatasourceName, workerNameLimit)
		task.EvalSlot = task.EvalSlot.UTC()
		normalized = append(normalized, task)
	}
	return normalized, nil
}

// truncateRunes 按字符截断，避免把 UTF-8 字符切坏。
func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

// utcTime 把可选时间统一成 UTC，空值保持为空。
func utcTime(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	utc := value.UTC()
	return &utc
}
