package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/worker"
)

// WorkerService exposes background worker telemetry to authorized administrators.
type WorkerService struct {
	v1.UnimplementedWorkerServiceServer
	workers *worker.Usecase
}

func NewWorkerService(uc *worker.Usecase) *WorkerService {
	return &WorkerService{workers: uc}
}

func (s *WorkerService) ListWorkerInstances(ctx context.Context, _ *v1.ListWorkerInstancesRequest) (*v1.ListWorkerInstancesResponse, error) {
	items, online, busy, idle, observedAt, err := s.workers.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.WorkerInstanceItem, 0, len(items))
	for _, item := range items {
		startedAt := ""
		if item.TaskStartedAt != nil {
			startedAt = item.TaskStartedAt.UTC().Format(time.RFC3339Nano)
		}
		out = append(out, &v1.WorkerInstanceItem{
			Service: item.Service, InstanceUid: item.InstanceUID, Hostname: item.Hostname,
			Busy: item.Busy, TaskUid: item.TaskUID, TaskName: item.TaskName,
			TaskStartedAt: startedAt, LastSeenAt: item.LastSeenAt.UTC().Format(time.RFC3339Nano), Online: item.Online,
		})
	}
	return &v1.ListWorkerInstancesResponse{
		Items: out, OnlineCount: int32(online), BusyCount: int32(busy), IdleCount: int32(idle),
		ObservedAt: observedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}
