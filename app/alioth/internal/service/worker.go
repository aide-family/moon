package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/biz/worker"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// WorkerService exposes background worker telemetry to authorized administrators.
type WorkerService struct {
	v1.UnimplementedWorkerServiceServer
	roster *worker.Roster
}

func NewWorkerService(roster *worker.Roster) *WorkerService {
	return &WorkerService{roster: roster}
}

func (s *WorkerService) ListWorkerInstances(ctx context.Context, _ *v1.ListWorkerInstancesRequest) (*v1.ListWorkerInstancesResponse, error) {
	items, online, busy, idle, observedAt, err := s.roster.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.WorkerInstanceItem, 0, len(items))
	for _, item := range items {
		out = append(out, workerInstanceItem(item))
	}
	return &v1.ListWorkerInstancesResponse{
		Items: out, OnlineCount: int32(online), BusyCount: int32(busy), IdleCount: int32(idle),
		ObservedAt: observedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

// GetWorkerInstance 返回单个实例的本批组合。倒计时需要两个时间点：
// 下一次玉衡发批由调度周期推出来，下一次本实例轮询由天权自己上报。
func (s *WorkerService) GetWorkerInstance(ctx context.Context, req *v1.GetWorkerInstanceRequest) (*v1.GetWorkerInstanceResponse, error) {
	item, observedAt, err := s.roster.Get(ctx, req.GetService(), req.GetInstanceUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetWorkerInstanceResponse{
		Item:                    workerInstanceItem(item),
		NextScheduleAt:          nextScheduleAt(observedAt).Format(time.RFC3339Nano),
		ScheduleIntervalSeconds: int32(alertevent.EvalInterval.Seconds()),
		ObservedAt:              observedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

// workerInstanceItem 把实例映射成接口返回项，含本批组合与幂等键。
func workerInstanceItem(item bo.WorkerInstance) *v1.WorkerInstanceItem {
	tasks := make([]*v1.WorkerTask, 0, len(item.Tasks))
	for _, task := range item.Tasks {
		tasks = append(tasks, &v1.WorkerTask{
			State: task.State, RuleUid: task.RuleUID, RuleName: task.RuleName,
			DatasourceUid: task.DatasourceUID, DatasourceName: task.DatasourceName,
			EvalSlot:       task.EvalSlot.UTC().Format(time.RFC3339Nano),
			IdempotencyKey: task.IdempotencyKey(),
		})
	}
	return &v1.WorkerInstanceItem{
		Service: item.Service, InstanceUid: item.InstanceUID, Hostname: item.Hostname,
		Busy: item.Busy, TaskUid: item.TaskUID, TaskName: item.TaskName,
		TaskStartedAt:       formatOptionalTime(item.TaskStartedAt),
		LastSeenAt:          item.LastSeenAt.UTC().Format(time.RFC3339Nano),
		Online:              item.Online,
		Tasks:               tasks,
		RoundStartedAt:      formatOptionalTime(item.RoundStartedAt),
		NextPollAt:          formatOptionalTime(item.NextPollAt),
		PollIntervalSeconds: item.PollIntervalSeconds,
	}
}

// nextScheduleAt 返回调度周期网格上的下一拍。玉衡的调度 cron 与墙钟对齐，
// 所以按周期向上取整即可，不需要等这一拍真的发生。
func nextScheduleAt(now time.Time) time.Time {
	if alertevent.EvalInterval <= 0 {
		return now.UTC()
	}
	return now.UTC().Truncate(alertevent.EvalInterval).Add(alertevent.EvalInterval)
}

func formatOptionalTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
