package worker

import (
	"context"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	recentWindow    = time.Minute
	onlineWindow    = 15 * time.Second
	recordRetention = 24 * time.Hour
)

var ProviderSet = wire.NewSet(NewUsecase)

// Usecase validates internal heartbeats and provides the admin worker overview.
type Usecase struct {
	repo   repository.WorkerRepo
	access access.Checker
	now    func() time.Time
}

func NewUsecase(repo repository.WorkerRepo, checker access.Checker) *Usecase {
	return &Usecase{repo: repo, access: checker, now: time.Now}
}

// Report stores one process heartbeat. The service name is fixed by the caller's internal gRPC endpoint.
func (uc *Usecase) Report(ctx context.Context, service string, status bo.WorkerStatus) error {
	service = strings.ToUpper(strings.TrimSpace(service))
	if service != "MEGREZ" && service != "MERAK" {
		return merr.Client("invalid worker service")
	}
	status.Service = service
	status.InstanceUID = strings.TrimSpace(status.InstanceUID)
	status.Hostname = strings.TrimSpace(status.Hostname)
	status.TaskUID = strings.TrimSpace(status.TaskUID)
	status.TaskName = strings.TrimSpace(status.TaskName)
	if status.InstanceUID == "" || len(status.InstanceUID) > 160 {
		return merr.Client("invalid worker instance")
	}
	if !status.Busy {
		status.TaskUID = ""
		status.TaskName = ""
		status.TaskStartedAt = nil
	} else if status.TaskUID == "" || len(status.TaskUID) > 160 || status.TaskStartedAt == nil || status.TaskStartedAt.IsZero() {
		return merr.Client("active task details are required")
	}
	if runes := []rune(status.TaskName); len(runes) > 240 {
		status.TaskName = string(runes[:240])
	}
	if runes := []rune(status.Hostname); len(runes) > 160 {
		status.Hostname = string(runes[:160])
	}
	return uc.repo.Report(ctx, status, uc.now().UTC())
}

// CleanupStale removes worker records that have exceeded the dashboard retention window.
func (uc *Usecase) CleanupStale(ctx context.Context) error {
	return uc.repo.CleanupBefore(ctx, uc.now().UTC().Add(-recordRetention))
}

func (uc *Usecase) List(ctx context.Context) ([]bo.WorkerInstance, int, int, int, time.Time, error) {
	if _, err := uc.access.RequireSystem(ctx, perm.SystemUserRead); err != nil {
		return nil, 0, 0, 0, time.Time{}, err
	}
	now := uc.now().UTC()
	items, err := uc.repo.ListRecent(ctx, now.Add(-recentWindow))
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
