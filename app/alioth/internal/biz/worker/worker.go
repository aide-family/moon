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
