package worker

import (
	"context"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

type fakeWorkerRepo struct {
	reported      bo.WorkerStatus
	items         []bo.WorkerInstance
	cleanedBefore time.Time
}

func (r *fakeWorkerRepo) Report(_ context.Context, status bo.WorkerStatus, _ time.Time) error {
	r.reported = status
	return nil
}

func (r *fakeWorkerRepo) ListRecent(context.Context, time.Time) ([]bo.WorkerInstance, error) {
	return r.items, nil
}

func (r *fakeWorkerRepo) CleanupBefore(_ context.Context, cutoff time.Time) error {
	r.cleanedBefore = cutoff
	return nil
}

type fakeWorkerUsers struct {
	repository.UserRepo
	account *bo.Account
}

func (r fakeWorkerUsers) UserByUID(context.Context, string) (*bo.Account, error) {
	return r.account, nil
}

func TestReportNormalizesIdleAndService(t *testing.T) {
	repo := &fakeWorkerRepo{}
	uc := NewUsecase(repo, access.Checker{})
	if err := uc.Report(context.Background(), " megrez ", bo.WorkerStatus{
		InstanceUID: "node-1", Busy: false, TaskUID: "stale-task", TaskName: "stale name",
	}); err != nil {
		t.Fatal(err)
	}
	if repo.reported.Service != "MEGREZ" || repo.reported.TaskUID != "" || repo.reported.TaskName != "" || repo.reported.TaskStartedAt != nil {
		t.Fatalf("idle heartbeat was not normalized: %+v", repo.reported)
	}
}

func TestCleanupStaleAppliesRetentionWindow(t *testing.T) {
	now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	repo := &fakeWorkerRepo{}
	uc := NewUsecase(repo, access.Checker{})
	uc.now = func() time.Time { return now }
	if err := uc.CleanupStale(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := now.Add(-recordRetention); !repo.cleanedBefore.Equal(want) {
		t.Fatalf("cleanup cutoff = %s, want %s", repo.cleanedBefore, want)
	}
}

func TestReportRejectsIncompleteBusyTask(t *testing.T) {
	uc := NewUsecase(&fakeWorkerRepo{}, access.Checker{})
	err := uc.Report(context.Background(), "MERAK", bo.WorkerStatus{InstanceUID: "node-1", Busy: true})
	if kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("expected client error for missing task details, got %v", err)
	}
}

func TestListCountsFreshWorkersAndRequiresSystemAccess(t *testing.T) {
	now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	repo := &fakeWorkerRepo{items: []bo.WorkerInstance{
		{Service: "MEGREZ", InstanceUID: "busy", Busy: true, LastSeenAt: now.Add(-5 * time.Second)},
		{Service: "MERAK", InstanceUID: "idle", LastSeenAt: now.Add(-10 * time.Second)},
		{Service: "MERAK", InstanceUID: "stale", Busy: true, LastSeenAt: now.Add(-30 * time.Second)},
	}}
	uc := NewUsecase(repo, access.Checker{Users: fakeWorkerUsers{account: &bo.Account{UID: "admin", Status: "ACTIVE", SystemRole: "SUPER_ADMIN"}}})
	uc.now = func() time.Time { return now }

	ctx := actor.WithUser(context.Background(), "admin")
	items, online, busy, idle, observedAt, err := uc.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if online != 2 || busy != 1 || idle != 1 || len(items) != 3 || !observedAt.Equal(now) {
		t.Fatalf("unexpected snapshot counts: online=%d busy=%d idle=%d items=%d observed=%s", online, busy, idle, len(items), observedAt)
	}
	if items[2].Online {
		t.Fatal("stale worker should not count as online")
	}
	if _, _, _, _, _, err := uc.List(context.Background()); kerrors.Reason(err) != merr.ReasonUnauthenticated {
		t.Fatalf("expected missing user to be rejected, got %v", err)
	}
}
