package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
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
	tasks, err := encodeWorkerTasks(status.Tasks)
	if err != nil {
		return err
	}
	_, err = r.deps.db.ExecContext(ctx, `
		INSERT INTO worker_instances
			(service, instance_uid, hostname, busy, task_uid, task_name, task_started_at,
			 tasks, round_started_at, next_poll_at, poll_interval_seconds, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (service, instance_uid) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			busy = EXCLUDED.busy,
			task_uid = EXCLUDED.task_uid,
			task_name = EXCLUDED.task_name,
			task_started_at = EXCLUDED.task_started_at,
			tasks = EXCLUDED.tasks,
			round_started_at = EXCLUDED.round_started_at,
			next_poll_at = EXCLUDED.next_poll_at,
			poll_interval_seconds = EXCLUDED.poll_interval_seconds,
			last_seen_at = EXCLUDED.last_seen_at`,
		status.Service, status.InstanceUID, status.Hostname, status.Busy,
		status.TaskUID, status.TaskName, status.TaskStartedAt,
		tasks, status.RoundStartedAt, status.NextPollAt, status.PollIntervalSeconds, now.UTC(),
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

func (r *workerRepo) Get(ctx context.Context, service, instanceUID string) (*bo.WorkerInstance, error) {
	var item bo.WorkerInstance
	var taskStartedAt, roundStartedAt, nextPollAt sql.NullTime
	var rawTasks string
	err := r.deps.db.QueryRowContext(ctx, `
		SELECT service, instance_uid, hostname, busy, task_uid, task_name, task_started_at,
			tasks, round_started_at, next_poll_at, poll_interval_seconds, last_seen_at
		FROM worker_instances
		WHERE service = $1 AND instance_uid = $2`, service, instanceUID).Scan(
		&item.Service, &item.InstanceUID, &item.Hostname, &item.Busy,
		&item.TaskUID, &item.TaskName, &taskStartedAt, &rawTasks,
		&roundStartedAt, &nextPollAt, &item.PollIntervalSeconds, &item.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merr.NotFound("worker instance not found")
		}
		r.log.Error(fmt.Sprintf("get worker instance: %v", err))
		return nil, databaseErr(r.log, err)
	}
	item.TaskStartedAt = nullableTime(taskStartedAt)
	item.RoundStartedAt = nullableTime(roundStartedAt)
	item.NextPollAt = nullableTime(nextPollAt)
	item.LastSeenAt = item.LastSeenAt.UTC()
	tasks, err := decodeWorkerTasks(rawTasks)
	if err != nil {
		return nil, err
	}
	item.Tasks = tasks
	return &item, nil
}

// encodeWorkerTasks 把组合快照写成 JSON 文本。快照只用于展示，不做查询条件，
// 所以整批覆盖比拆表更简单，也不会随组合数量增长改表结构。
func encodeWorkerTasks(tasks []bo.WorkerTask) (string, error) {
	if len(tasks) == 0 {
		return "[]", nil
	}
	encoded, err := json.Marshal(tasks)
	if err != nil {
		slog.Default().Error(fmt.Sprintf("encode worker tasks: %v", err))
		return "", merr.Server("request failed")
	}
	return string(encoded), nil
}

func decodeWorkerTasks(raw string) ([]bo.WorkerTask, error) {
	if raw == "" {
		return nil, nil
	}
	tasks := make([]bo.WorkerTask, 0)
	if err := json.Unmarshal([]byte(raw), &tasks); err != nil {
		slog.Default().Error(fmt.Sprintf("decode worker tasks: %v", err))
		return nil, merr.Server("request failed")
	}
	return tasks, nil
}

// nullableTime 把可空时间列转成指针，空值表示这个时间还没有。
func nullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	out := value.Time.UTC()
	return &out
}
