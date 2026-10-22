package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

var _ repository.WorkerRepo = (*workerRepo)(nil)

func NewWorkerRepo(deps *Dependencies, logger *slog.Logger) repository.WorkerRepo {
	return &workerRepo{deps: deps, log: logger}
}

type workerRepo struct {
	deps *Dependencies
	log  *slog.Logger
}

func (r *workerRepo) Report(ctx context.Context, status bo.WorkerStatus, now time.Time) error {
	_, err := r.deps.db.ExecContext(ctx, `
		INSERT INTO worker_instances
			(service, instance_uid, hostname, busy, task_uid, task_name, task_started_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (service, instance_uid) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			busy = EXCLUDED.busy,
			task_uid = EXCLUDED.task_uid,
			task_name = EXCLUDED.task_name,
			task_started_at = EXCLUDED.task_started_at,
			last_seen_at = EXCLUDED.last_seen_at`,
		status.Service, status.InstanceUID, status.Hostname, status.Busy,
		status.TaskUID, status.TaskName, status.TaskStartedAt, now.UTC(),
	)
	if err != nil {
		r.log.Error(fmt.Sprintf("worker heartbeat: %v", err))
		return databaseErr(r.log, err)
	}
	return nil
}

func (r *workerRepo) CleanupBefore(ctx context.Context, cutoff time.Time) error {
	if _, err := r.deps.db.ExecContext(ctx, `DELETE FROM worker_instances WHERE last_seen_at < $1`, cutoff.UTC()); err != nil {
		r.log.Error(fmt.Sprintf("cleanup stale worker instances: %v", err))
		return databaseErr(r.log, err)
	}
	return nil
}

func (r *workerRepo) ListRecent(ctx context.Context, since time.Time) ([]bo.WorkerInstance, error) {
	rows, err := r.deps.db.QueryContext(ctx, `
		SELECT service, instance_uid, hostname, busy, task_uid, task_name, task_started_at, last_seen_at
		FROM worker_instances
		WHERE last_seen_at >= $1
		ORDER BY service, busy DESC, last_seen_at DESC, hostname, instance_uid`, since.UTC())
	if err != nil {
		r.log.Error(fmt.Sprintf("list worker instances: %v", err))
		return nil, databaseErr(r.log, err)
	}
	defer rows.Close()

	items := make([]bo.WorkerInstance, 0)
	for rows.Next() {
		var item bo.WorkerInstance
		var taskStartedAt sql.NullTime
		if err := rows.Scan(
			&item.Service, &item.InstanceUID, &item.Hostname, &item.Busy,
			&item.TaskUID, &item.TaskName, &taskStartedAt, &item.LastSeenAt,
		); err != nil {
			r.log.Error(fmt.Sprintf("scan worker instance: %v", err))
			return nil, databaseErr(r.log, err)
		}
		if taskStartedAt.Valid {
			started := taskStartedAt.Time.UTC()
			item.TaskStartedAt = &started
		}
		item.LastSeenAt = item.LastSeenAt.UTC()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		r.log.Error(fmt.Sprintf("iterate worker instances: %v", err))
		return nil, databaseErr(r.log, err)
	}
	return items, nil
}
