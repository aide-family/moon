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
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulenotificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
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

	ns := "ns-notify-" + time.Now().Format("150405.000")
	notify := NewNotificationRepo(store, slog.Default())
	events := NewAlertEventRepo(store, slog.Default())
	rules := NewAlertRuleRepo(store, slog.Default())
	t.Cleanup(func() {
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertRuleNotificationGroup.Delete().Where(alertrulenotificationgroup.NamespaceUIDEQ(ns)).Exec(ctx)
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
	g, err := rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: "g1"})
	if err != nil {
		t.Fatal(err)
	}
	rule, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: g.UID, Name: "r1", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Expr: "up", NotificationGroupUIDs: []string{group.UID},
		Levels: []bo.AlertRuleLevel{{Condition: "EQ", Sample: "MAX", Threshold: "1", LevelUID: "lv"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rule.NotificationGroupUIDs) != 1 || rule.NotificationGroupUIDs[0] != group.UID {
		t.Fatalf("bound groups = %v", rule.NotificationGroupUIDs)
	}

	slot := time.Now().UTC().Truncate(time.Minute)
	if err := events.ApplyReport(ctx, *rule, bo.EvalReport{
		RuleUID: rule.UID, EvalSlot: slot,
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
		RuleUID: rule.UID, EvalSlot: slot.Add(time.Minute), Hits: nil,
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
		Expr: "up", NotificationGroupUIDs: []string{group.UID},
		Levels: []bo.AlertRuleLevel{{Condition: "EQ", Sample: "MAX", Threshold: "1", LevelUID: "lv"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := events.ApplyReport(ctx, *rule2, bo.EvalReport{
		RuleUID: rule2.UID, EvalSlot: slot.Add(2 * time.Minute),
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
}
