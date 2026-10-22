package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestWorkerRepoHeartbeatUpsert(t *testing.T) {
	ctx := context.Background()
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: testDSN(t)}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"}, slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	defer cleanup()
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}

	repo := NewWorkerRepo(store, slog.Default()).(*workerRepo)
	instanceUID := "test-worker-" + time.Now().UTC().Format("150405.000000000")
	staleUID := instanceUID + "-stale"
	defer func() {
		_, _ = store.db.ExecContext(ctx, `DELETE FROM worker_instances WHERE instance_uid IN ($1, $2)`, instanceUID, staleUID)
	}()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO worker_instances (service, instance_uid, busy, task_uid, task_name, last_seen_at)
		VALUES ($1, $2, false, '', '', $3)
	`, "MERAK", staleUID, now.Add(-25*time.Hour)); err != nil {
		t.Fatal(err)
	}
	startedAt := now.Add(-time.Minute)
	status := bo.WorkerStatus{
		Service: "MEGREZ", InstanceUID: instanceUID, Hostname: "worker-test", Busy: true,
		TaskUID: "rule-1", TaskName: "CPU alert", TaskStartedAt: &startedAt,
	}
	if err := repo.Report(ctx, status, now); err != nil {
		t.Fatal(err)
	}
	status.Busy = false
	status.TaskUID = ""
	status.TaskName = ""
	status.TaskStartedAt = nil
	if err := repo.Report(ctx, status, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := repo.CleanupBefore(ctx, now.Add(-24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListRecent(ctx, now.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	var staleCount int
	if err := store.db.QueryRowContext(ctx, `SELECT count(*) FROM worker_instances WHERE instance_uid = $1`, staleUID).Scan(&staleCount); err != nil {
		t.Fatal(err)
	}
	if staleCount != 0 {
		t.Fatalf("stale worker record count = %d, want 0", staleCount)
	}
	for _, item := range items {
		if item.InstanceUID != instanceUID {
			continue
		}
		if item.Service != "MEGREZ" || item.Busy || item.TaskUID != "" || item.TaskName != "" || item.TaskStartedAt != nil {
			t.Fatalf("heartbeat upsert retained stale task state: %+v", item)
		}
		if !item.LastSeenAt.Equal(now.Add(time.Second)) {
			t.Fatalf("last seen = %s, want %s", item.LastSeenAt, now.Add(time.Second))
		}
		return
	}
	t.Fatal("upserted worker heartbeat was not listed")
}
