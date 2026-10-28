package worker

import (
	"context"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
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

func (r *fakeWorkerRepo) Get(_ context.Context, service, instanceUID string) (*bo.WorkerInstance, error) {
	for i := range r.items {
		if r.items[i].Service == service && r.items[i].InstanceUID == instanceUID {
			return &r.items[i], nil
		}
	}
	return nil, merr.NotFound("worker instance not found")
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
	roster := NewRoster(repo, access.Checker{})
	if err := roster.Report(context.Background(), " megrez ", bo.WorkerStatus{
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
	roster := NewRoster(repo, access.Checker{})
	roster.now = func() time.Time { return now }
	if err := roster.CleanupStale(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := now.Add(-recordRetention); !repo.cleanedBefore.Equal(want) {
		t.Fatalf("cleanup cutoff = %s, want %s", repo.cleanedBefore, want)
	}
}

func TestReportRejectsIncompleteBusyTask(t *testing.T) {
	roster := NewRoster(&fakeWorkerRepo{}, access.Checker{})
	err := roster.Report(context.Background(), "MERAK", bo.WorkerStatus{InstanceUID: "node-1", Busy: true})
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
	roster := NewRoster(repo, access.Checker{Users: fakeWorkerUsers{account: &bo.Account{UID: "admin", Status: "ACTIVE", SystemRole: "SUPER_ADMIN"}}})
	roster.now = func() time.Time { return now }

	ctx := actor.WithUser(context.Background(), "admin")
	items, online, busy, idle, observedAt, err := roster.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if online != 2 || busy != 1 || idle != 1 || len(items) != 3 || !observedAt.Equal(now) {
		t.Fatalf("unexpected snapshot counts: online=%d busy=%d idle=%d items=%d observed=%s", online, busy, idle, len(items), observedAt)
	}
	if items[2].Online {
		t.Fatal("stale worker should not count as online")
	}
	if _, _, _, _, _, err := roster.List(context.Background()); kerrors.Reason(err) != merr.ReasonUnauthenticated {
		t.Fatalf("expected missing user to be rejected, got %v", err)
	}
}

func TestReportNormalizesWorkerTasks(t *testing.T) {
	repo := &fakeWorkerRepo{}
	roster := NewRoster(repo, access.Checker{})
	slot := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	err := roster.Report(context.Background(), "MEGREZ", bo.WorkerStatus{
		InstanceUID: "node-1",
		Tasks: []bo.WorkerTask{
			{State: " running ", RuleUID: " rule-1 ", RuleName: " CPU ", DatasourceUID: "ds-a", DatasourceName: "Prom", EvalSlot: slot},
			{State: bo.WorkerTaskNew, RuleUID: "rule-1", RuleName: "duplicate", DatasourceUID: "ds-a", EvalSlot: slot},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	tasks := repo.reported.Tasks
	if len(tasks) != 1 {
		t.Fatalf("duplicate combination kept: %+v", tasks)
	}
	if tasks[0].State != bo.WorkerTaskRunning || tasks[0].RuleUID != "rule-1" || tasks[0].RuleName != "CPU" {
		t.Fatalf("task was not normalized: %+v", tasks[0])
	}
	if !tasks[0].EvalSlot.Equal(slot) {
		t.Fatalf("eval slot = %s, want %s", tasks[0].EvalSlot, slot)
	}
}

func TestReportRejectsUnknownTaskState(t *testing.T) {
	roster := NewRoster(&fakeWorkerRepo{}, access.Checker{})
	err := roster.Report(context.Background(), "MEGREZ", bo.WorkerStatus{
		InstanceUID: "node-1",
		Tasks:       []bo.WorkerTask{{State: "SLEEPING", RuleUID: "rule-1", DatasourceUID: "ds-a", EvalSlot: time.Now()}},
	})
	if kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("expected client error for unknown state, got %v", err)
	}
}

func TestGetReturnsTasksAndRequiresSystemAccess(t *testing.T) {
	now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	slot := now
	repo := &fakeWorkerRepo{items: []bo.WorkerInstance{{
		Service: "MEGREZ", InstanceUID: "node-1", Hostname: "node-1", Busy: true,
		TaskUID: "rule-1", TaskName: "CPU", TaskStartedAt: &now, LastSeenAt: now.Add(-5 * time.Second),
		Tasks: []bo.WorkerTask{{
			State: bo.WorkerTaskRunning, RuleUID: "rule-1", RuleName: "CPU",
			DatasourceUID: "ds-a", DatasourceName: "Prom", EvalSlot: slot,
		}},
	}}}
	roster := NewRoster(repo, access.Checker{Users: fakeWorkerUsers{account: &bo.Account{UID: "admin", Status: "ACTIVE", SystemRole: "SUPER_ADMIN"}}})
	roster.now = func() time.Time { return now }

	ctx := actor.WithUser(context.Background(), "admin")
	item, observedAt, err := roster.Get(ctx, " megrez ", " node-1 ")
	if err != nil {
		t.Fatal(err)
	}
	if item.Service != "MEGREZ" || !item.Online || !observedAt.Equal(now) || len(item.Tasks) != 1 {
		t.Fatalf("detail = %+v observed=%s", item, observedAt)
	}
	if key, want := item.Tasks[0].IdempotencyKey(), "rule-1/ds-a@2026-10-21T08:00:00Z"; key != want {
		t.Fatalf("idempotency key = %s, want %s", key, want)
	}

	if _, _, err := roster.Get(context.Background(), "MEGREZ", "node-1"); kerrors.Reason(err) != merr.ReasonUnauthenticated {
		t.Fatalf("expected missing user to be rejected, got %v", err)
	}
	if _, _, err := roster.Get(ctx, "ALKaid", "node-1"); kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("expected unknown service to be rejected, got %v", err)
	}
	if _, _, err := roster.Get(ctx, "MEGREZ", ""); kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("expected empty instance to be rejected, got %v", err)
	}
}
