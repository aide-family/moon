package data

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/messagelog"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/messagelogshard"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// TestNotificationListFilters 覆盖投递日志和通知任务的状态、任务与时间范围筛选，
// 同时验证归档到租户周表的日志仍能参与列表和详情查询。
func TestNotificationListFilters(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}

	ns := "ns-filter-" + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		_, _ = store.client.MessageLogShard.Delete().Where(messagelogshard.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.MessageLog.Delete().Where(messagelog.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(ns)).Exec(ctx)
		_ = store.shards.dropTenant(context.Background(), ns)
	})

	repo := NewNotificationRepo(store, slog.Default()).(*notificationRepo)
	base := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	taskUID := "task-" + ns
	tasks := []struct {
		status notificationtask.Status
		at     time.Time
	}{
		{notificationtask.StatusSENT, base.Add(time.Hour)},
		{notificationtask.StatusFAILED, base.Add(3 * time.Hour)},
	}
	for i, item := range tasks {
		if _, err := store.client.NotificationTask.Create().
			SetUID(fmt.Sprintf("task-%d-%s", i, ns)).
			SetNamespaceUID(ns).
			SetChannelUID("channel").
			SetStatus(item.status).
			SetCreatedAt(item.at).
			Save(ctx); err != nil {
			t.Fatal(err)
		}
	}

	logs := []struct {
		status messagelog.Status
		at     time.Time
	}{
		{messagelog.StatusSENT, base},
		{messagelog.StatusFAILED, base.Add(2 * time.Hour)},
		{messagelog.StatusSENT, base.Add(4 * time.Hour)},
	}
	logUIDs := make([]string, 0, len(logs))
	for i, item := range logs {
		uid := fmt.Sprintf("log-%d-%s", i, ns)
		logUIDs = append(logUIDs, uid)
		if _, err := store.client.MessageLog.Create().
			SetUID(uid).
			SetNamespaceUID(ns).
			SetTaskUID(taskUID).
			SetStatus(item.status).
			SetCreatedAt(item.at).
			Save(ctx); err != nil {
			t.Fatal(err)
		}
	}
	// 第一条日志走归档路径，其余留在主库兜底，列表需要合并两个来源。
	repo.archiveMessageLog(ctx, bo.MessageLog{
		UID: logUIDs[0], NamespaceUID: ns, TaskUID: taskUID,
		Status: string(messagelog.StatusSENT), CreatedAt: base,
	})

	taskItems, total, err := repo.ListTasks(ctx, ns, bo.NotificationTaskListFilter{Status: "SENT"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(taskItems) != 1 || taskItems[0].UID != fmt.Sprintf("task-0-%s", ns) {
		t.Fatalf("tasks by status = %d %+v", total, taskItems)
	}
	_, total, err = repo.ListTasks(ctx, ns, bo.NotificationTaskListFilter{
		StartAt: base.Add(2 * time.Hour), EndAt: base.Add(4 * time.Hour),
	}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("tasks by time range = %d", total)
	}

	items, total, err := repo.ListMessageLogs(ctx, ns, bo.MessageLogListFilter{Status: "SENT"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(items) != 2 || items[0].UID != logUIDs[2] || items[1].UID != logUIDs[0] {
		t.Fatalf("logs by status = %d %+v", total, items)
	}
	items, total, err = repo.ListMessageLogs(ctx, ns, bo.MessageLogListFilter{
		StartAt: base.Add(time.Hour), EndAt: base.Add(3 * time.Hour),
	}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].UID != logUIDs[1] {
		t.Fatalf("logs by time range = %d %+v", total, items)
	}
	items, total, err = repo.ListMessageLogs(ctx, ns, bo.MessageLogListFilter{TaskUID: taskUID}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(items) != 3 {
		t.Fatalf("logs by task = %d %+v", total, items)
	}

	// 归档后的详情按分片索引定位周表。
	item, err := repo.GetMessageLog(ctx, ns, logUIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != "SENT" {
		t.Fatalf("archived log = %+v", item)
	}
}
