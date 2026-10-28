package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestDatasourceAvailabilityTriggersAndRecovers 校验失败批次按窗口累计、
// 达到阈值生成一条数据源可用性事件，并在下一次成功评测后恢复归档。
// 需要 ALIOTH_TEST_DSN 指向非 dev 库；没有配置时跳过。
func TestDatasourceAvailabilityTriggersAndRecovers(t *testing.T) {
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
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	namespace := "ns-availability-" + time.Now().Format("150405.000000")
	sources := NewDatasourceRepo(store, slog.Default())
	created, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: namespace, Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: "http://127.0.0.1:9090", Labels: `{"cluster":"prod"}`,
		AvailabilityAlertEnabled: true, FailureWindowMinutes: 5, FailureThreshold: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = store.client.Datasource.Delete().Where(datasource.UIDEQ(created.UID)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.DatasourceUIDEQ(created.UID)).Exec(ctx)
		_, _ = store.client.AlertEventFiring.Delete().Where(alerteventfiring.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationChannel.Delete().Where(notificationchannel.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationTemplate.Delete().Where(notificationtemplate.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationGroup.Delete().Where(notificationgroup.NamespaceUIDEQ(namespace)).Exec(ctx)
		_ = store.history.dropTenant(context.Background(), namespace)
	})

	// 绑定一个可投递的邮件通知组，验证条件事件会生成 FIRING 与恢复任务。
	channel, err := store.client.NotificationChannel.Create().
		SetUID("ch-" + namespace).SetNamespaceUID(namespace).SetName("ops-mail").
		SetType(notificationchannel.TypeSMTP).SetEndpoint("smtp.example.com:465").
		SetFromAddress("alerts@example.com").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	template, err := store.client.NotificationTemplate.Create().
		SetUID("tpl-" + namespace).SetNamespaceUID(namespace).SetName("ops-mail").
		SetType(notificationtemplate.TypeEMAIL).SetSubject("alert").SetBody("{{ .Summary }}").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	group, err := store.client.NotificationGroup.Create().
		SetUID("grp-" + namespace).SetNamespaceUID(namespace).SetName("ops").
		SetChannelUids([]string{channel.UID}).SetTemplateUids([]string{template.UID}).
		SetEmailReceivers("ops@example.com").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.client.Datasource.Update().Where(datasource.UIDEQ(created.UID)).
		SetNotificationGroupUids([]string{group.UID}).Save(ctx); err != nil {
		t.Fatal(err)
	}

	events := NewAlertEventRepo(store, slog.Default())
	rule := bo.AlertRule{
		UID: "rule-" + namespace, NamespaceUID: namespace, DatasourceUID: created.UID,
		Expr: "up", Summary: "cpu",
	}
	slot := time.Now().UTC().Truncate(time.Second)
	report := func(failed bool, offset int) {
		t.Helper()
		if err := events.ApplyReport(ctx, rule, bo.EvalReport{
			RuleUID: rule.UID, DatasourceUID: created.UID, EvalSlot: slot.Add(time.Duration(offset) * time.Second),
			Failed: failed, FailureMessage: "upstream query failed",
		}); err != nil {
			t.Fatal(err)
		}
	}
	for offset := 0; offset < 2; offset++ {
		report(true, offset)
	}
	if count := firingEvents(t, store, namespace); len(count) != 0 {
		t.Fatalf("event fired before threshold: %+v", count)
	}
	report(true, 2)
	items := firingEvents(t, store, namespace)
	if len(items) != 1 {
		t.Fatalf("availability events = %+v", items)
	}
	item := items[0]
	if item.RuleUID != "" || item.DatasourceUID != created.UID || item.SampleValue != "3" {
		t.Fatalf("availability event = %+v", item)
	}
	if item.Summary == "" || item.Description == "" || item.Fingerprint == "" {
		t.Fatalf("availability event text = %+v", item)
	}
	tasks, err := store.client.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(item.UID),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Phase != notificationtask.PhaseFIRING || tasks[0].GroupUID != group.UID ||
		tasks[0].ChannelUID != channel.UID || tasks[0].Receivers != "ops@example.com" ||
		tasks[0].Fingerprint != item.Fingerprint || tasks[0].RuleUID != "" {
		t.Fatalf("firing tasks = %+v", tasks)
	}

	// 再次失败只刷新一条事件，不重复创建，也不重复投递 FIRING 通知。
	report(true, 3)
	if items := firingEvents(t, store, namespace); len(items) != 1 || items[0].SampleValue != "4" {
		t.Fatalf("event after another failure = %+v", items)
	}
	if fired, err := store.client.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(item.UID),
		notificationtask.PhaseEQ(notificationtask.PhaseFIRING),
	).Count(ctx); err != nil || fired != 1 {
		t.Fatalf("firing tasks after another failure = %d, err = %v", fired, err)
	}

	// 成功评测收口事件并归档到历史分片，主库不再保留，FIRING 占位也删除。
	report(false, 4)
	if items := firingEvents(t, store, namespace); len(items) != 0 {
		t.Fatalf("event still firing after success: %+v", items)
	}
	if firing := store.client.AlertEventFiring.Query().Where(alerteventfiring.NamespaceUIDEQ(namespace)).CountX(ctx); firing != 0 {
		t.Fatalf("firing placeholders = %d", firing)
	}
	history, _, err := events.ListEvents(ctx, namespace, bo.AlertEventListFilter{View: "HISTORY"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Status != "RECOVERED" || history[0].EndsAt == nil {
		t.Fatalf("history = %+v", history)
	}
	if history[0].RuleUID != "" {
		t.Fatalf("availability history rule uid = %q", history[0].RuleUID)
	}
	recoveryTasks, err := store.client.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(item.UID),
		notificationtask.PhaseEQ(notificationtask.PhaseRECOVERED),
	).All(ctx)
	if err != nil || len(recoveryTasks) != 1 || recoveryTasks[0].GroupUID != group.UID {
		t.Fatalf("recovery tasks = %+v, err = %v", recoveryTasks, err)
	}

	// 失败批次记为已回写，成功批次也存在；失败批次只统计 4 条，且没有误恢复规则事件。
	failed := store.client.AlertEvalBatch.Query().Where(
		alertevalbatch.DatasourceUIDEQ(created.UID),
		alertevalbatch.FailedEQ(true),
	).CountX(ctx)
	if failed != 4 {
		t.Fatalf("failed batches = %d", failed)
	}
}

func firingEvents(t *testing.T, store *Dependencies, namespace string) []bo.AlertEvent {
	t.Helper()
	rows, err := store.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespace),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	items := make([]bo.AlertEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.AlertEvent{
			UID: row.UID, RuleUID: row.RuleUID, DatasourceUID: row.DatasourceUID,
			Fingerprint: row.Fingerprint, SampleValue: row.SampleValue,
			Summary: row.Summary, Description: row.Description,
		})
	}
	return items
}
