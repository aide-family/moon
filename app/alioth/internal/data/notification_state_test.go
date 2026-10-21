package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/messagelog"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// TestNotificationStateCAS 锁定通知任务的状态机语义：
// 只有 FAILED 能重试、只有 PENDING/FAILED 能取消、只有 SENDING 接受回写，
// 且重复或迟到的回写不能产生第二条消息日志。
func TestNotificationStateCAS(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewData(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	defer cleanup()
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}

	ns := "ns-state-" + time.Now().Format("150405.000")
	repo := NewNotificationRepo(store, slog.Default())
	uid := "task-state-" + ns
	if _, err := store.client.NotificationTask.Create().
		SetUID(uid).
		SetNamespaceUID(ns).
		SetChannelUID("channel-1").
		SetStatus(notificationtask.StatusPENDING).
		SetReceivers("ops@example.com").
		Save(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = store.client.MessageLog.Delete().Where(messagelog.TaskUIDEQ(uid)).Exec(ctx)
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.UIDEQ(uid)).Exec(ctx)
	})

	logCount := func() int {
		n, err := store.client.MessageLog.Query().Where(messagelog.TaskUIDEQ(uid)).Count(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	setStatus := func(status notificationtask.Status) {
		if _, err := store.db.ExecContext(ctx,
			`UPDATE notification_tasks SET status = $2, retries_left = 3, updated_at = now() WHERE uid = $1`,
			uid, string(status),
		); err != nil {
			t.Fatal(err)
		}
	}

	// PENDING 不能重试，也不能接受回写。
	if _, err := repo.RetryTask(ctx, ns, uid); err == nil {
		t.Fatal("retry on PENDING must fail")
	}
	if err := repo.ApplyNotifyReport(ctx, bo.NotifyReport{TaskUID: uid, OK: true}); err != nil {
		t.Fatal(err)
	}
	if logCount() != 0 {
		t.Fatal("report on non-SENDING task must not write a message log")
	}

	// SENDING 不能取消；但可以回写一次。
	setStatus(notificationtask.StatusSENDING)
	if _, err := repo.CancelTask(ctx, ns, uid); err == nil {
		t.Fatal("cancel on SENDING must fail")
	}
	if err := repo.ApplyNotifyReport(ctx, bo.NotifyReport{TaskUID: uid, OK: true}); err != nil {
		t.Fatal(err)
	}
	row, err := store.client.NotificationTask.Query().Where(notificationtask.UIDEQ(uid)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != notificationtask.StatusSENT {
		t.Fatalf("status = %s, want SENT", row.Status)
	}
	if logCount() != 1 {
		t.Fatalf("message logs = %d, want 1", logCount())
	}

	// 迟到/重复的回写不能把 SENT 退回 FAILED，也不能再写日志。
	if err := repo.ApplyNotifyReport(ctx, bo.NotifyReport{TaskUID: uid, OK: false, Detail: "late"}); err != nil {
		t.Fatal(err)
	}
	row, err = store.client.NotificationTask.Query().Where(notificationtask.UIDEQ(uid)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != notificationtask.StatusSENT {
		t.Fatalf("late report changed status to %s", row.Status)
	}
	if logCount() != 1 {
		t.Fatalf("late report wrote a second message log: %d", logCount())
	}

	// FAILED 才能重试；重试后回到 PENDING 并清空 next_attempt_at。
	setStatus(notificationtask.StatusFAILED)
	retried, err := repo.RetryTask(ctx, ns, uid)
	if err != nil {
		t.Fatal(err)
	}
	if retried.Status != "PENDING" || retried.RetriesLeft != 3 || retried.LastError != "" {
		t.Fatalf("retried task = %+v", retried)
	}

	// PENDING 可以取消。
	if _, err := repo.CancelTask(ctx, ns, uid); err != nil {
		t.Fatal(err)
	}
	row, err = store.client.NotificationTask.Query().Where(notificationtask.UIDEQ(uid)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != notificationtask.StatusCANCELLED {
		t.Fatalf("status = %s, want CANCELLED", row.Status)
	}

	// 超时回收后迟到的回写同样不能改状态或追加日志。
	setStatus(notificationtask.StatusSENDING)
	if _, err := store.db.ExecContext(ctx,
		`UPDATE notification_tasks SET updated_at = now() - interval '10 minutes' WHERE uid = $1`, uid,
	); err != nil {
		t.Fatal(err)
	}
	reaped, err := repo.ReapStaleSending(ctx, time.Now().Add(-2*time.Minute), 100)
	if err != nil {
		t.Fatal(err)
	}
	if reaped == 0 {
		t.Fatal("stale task was not reaped")
	}
	if err := repo.ApplyNotifyReport(ctx, bo.NotifyReport{TaskUID: uid, OK: true}); err != nil {
		t.Fatal(err)
	}
	row, err = store.client.NotificationTask.Query().Where(notificationtask.UIDEQ(uid)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != notificationtask.StatusFAILED {
		t.Fatalf("late report revived reaped task: %s", row.Status)
	}
	if logCount() != 2 {
		t.Fatalf("message logs = %d, want 2 (manual + timeout)", logCount())
	}
}
