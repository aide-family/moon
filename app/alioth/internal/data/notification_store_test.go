package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulegroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestNotificationTaskDedup(t *testing.T) {
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
	if err := seedNotificationPresets(ctx, store.client, slog.Default()); err != nil {
		t.Fatal(err)
	}

	ns := "ns-notify-" + time.Now().Format("150405.000")
	notify := NewNotificationRepo(store, slog.Default())
	events := NewAlertEventRepo(store, slog.Default())
	rules := NewAlertRuleRepo(store, slog.Default())
	t.Cleanup(func() {
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationSubscriptionGroup.Delete().Where(notificationsubscriptiongroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationSubscription.Delete().Where(notificationsubscription.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertRuleLevel.Delete().Where(alertrulelevel.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertRule.Delete().Where(alertrule.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertRuleGroup.Delete().Where(alertrulegroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationGroup.Delete().Where(notificationgroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationTemplate.Delete().Where(notificationtemplate.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationChannel.Delete().Where(notificationchannel.NamespaceUIDEQ(ns)).Exec(ctx)
	})

	channel, err := notify.CreateChannel(ctx, bo.NotificationChannel{
		NamespaceUID: ns, Name: "smtp", Type: "SMTP", App: "OTHER",
		Endpoint: "127.0.0.1:1025", Password: "secret", FromAddress: "a@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	row, err := store.client.NotificationChannel.Query().Where(notificationchannel.UIDEQ(channel.UID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.PasswordCipher == "secret" || row.PasswordCipher == "" {
		t.Fatal("channel password is not ciphertext")
	}
	if channel.HasPassword && channel.Password != "" {
		t.Fatal("channel password leaked in API shape")
	}

	template, err := notify.CreateTemplate(ctx, bo.NotificationTemplate{
		NamespaceUID: ns, Name: "tpl", Type: "EMAIL", Subject: "hi", Body: "{{.summary}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := notify.CreateGroup(ctx, bo.NotificationGroup{
		NamespaceUID: ns, Name: "ops", ChannelUID: channel.UID, TemplateUID: template.UID, Receivers: "ops@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "sub", Labels: "{}", GroupUIDs: []string{group.UID},
	}); err != nil {
		t.Fatal(err)
	}
	g, err := rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: "g1"})
	if err != nil {
		t.Fatal(err)
	}
	rule, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: g.UID, Name: "r1", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Expr:   "up",
		Levels: []bo.AlertRuleLevel{{Condition: "EQ", Sample: "MAX", Threshold: "1", LevelUID: "lv"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	slot := time.Now().UTC().Truncate(time.Minute)
	if err := events.ApplyReport(ctx, *rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: "ds1", EvalSlot: slot,
		Hits: []bo.EvalHit{{Fingerprint: "fp1", Labels: "{}", Value: "1", LevelUID: "lv", Summary: "fire"}},
	}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err := notify.ListTasks(ctx, ns, "", 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Phase != "FIRING" {
		t.Fatalf("firing tasks = %+v", tasks)
	}

	if err := events.ApplyReport(ctx, *rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: "ds1", EvalSlot: slot.Add(time.Minute), Hits: nil,
	}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err = notify.ListTasks(ctx, ns, "", 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("want firing+recovered, got %d", len(tasks))
	}
	phases := map[string]int{}
	for _, task := range tasks {
		phases[task.Phase]++
	}
	if phases["FIRING"] != 1 || phases["RECOVERED"] != 1 {
		t.Fatalf("phases = %v", phases)
	}

	if _, err := notify.UpdateGroupStatus(ctx, ns, group.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	before := len(tasks)
	rule2, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: g.UID, Name: "r2", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Expr:   "up",
		Levels: []bo.AlertRuleLevel{{Condition: "EQ", Sample: "MAX", Threshold: "1", LevelUID: "lv"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := events.ApplyReport(ctx, *rule2, bo.EvalReport{
		RuleUID: rule2.UID, DatasourceUID: "ds1", EvalSlot: slot.Add(2 * time.Minute),
		Hits: []bo.EvalHit{{Fingerprint: "fp2", Labels: "{}", Value: "1", LevelUID: "lv"}},
	}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err = notify.ListTasks(ctx, ns, "", 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != before {
		t.Fatalf("disabled group still created tasks: before=%d after=%d", before, len(tasks))
	}

	task, err := notify.CreateTestTask(ctx, bo.NotificationTask{
		NamespaceUID: ns, ChannelUID: channel.UID, TemplateUID: template.UID,
		Receivers: "test@example.com", Subject: "t", Body: "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != "PENDING" || task.Phase != "TEST" {
		t.Fatalf("test task = %+v", task)
	}

	// 退避：next_attempt_at 在未来的任务不能被领取；到点后恢复可领取。
	backoffUID := "task-backoff-" + ns
	if _, err := store.client.NotificationTask.Create().
		SetUID(backoffUID).
		SetNamespaceUID(ns).
		SetChannelUID(channel.UID).
		SetStatus(notificationtask.StatusPENDING).
		SetNextAttemptAt(time.Now().UTC().Add(time.Minute)).
		SetReceivers("backoff@example.com").
		Save(ctx); err != nil {
		t.Fatal(err)
	}
	repo := NewNotificationRepo(store, slog.Default())
	pending, err := repo.ClaimPending(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range pending {
		if item.UID == backoffUID {
			t.Fatal("task with future next_attempt_at was claimed")
		}
	}
	if _, err := store.client.NotificationTask.Update().
		Where(notificationtask.UIDEQ(backoffUID)).
		SetNextAttemptAt(time.Now().UTC().Add(-time.Minute)).
		Save(ctx); err != nil {
		t.Fatal(err)
	}
	pending, err = repo.ClaimPending(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range pending {
		if item.UID == backoffUID {
			found = true
		}
	}
	if !found {
		t.Fatal("task with expired next_attempt_at was not claimed")
	}

	// SENDING 超时回收：模拟 merak 改为 SENDING 后崩溃。
	if _, err := store.db.ExecContext(ctx,
		`UPDATE notification_tasks SET status = 'SENDING', updated_at = now() - interval '10 minutes' WHERE uid = $1`,
		task.UID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReapStaleSending(ctx, time.Now().Add(-2*time.Minute), 10); err != nil {
		t.Fatal(err)
	}
	stale, err := store.client.NotificationTask.Query().Where(notificationtask.UIDEQ(task.UID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if stale.Status != notificationtask.StatusFAILED || stale.LastError != "delivery timed out" {
		t.Fatalf("stale task = %s / %s", stale.Status, stale.LastError)
	}

	presets, err := repo.ListPresets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) == 0 {
		t.Fatal("no preset templates seeded")
	}
}
