package data

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestNotificationSubscriptionFanout(t *testing.T) {
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
	defer cleanup()
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	ns := "ns-sub-" + time.Now().Format("150405.000")
	t.Cleanup(func() {
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationSubscriptionGroup.Delete().Where(notificationsubscriptiongroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationSubscription.Delete().Where(notificationsubscription.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationGroup.Delete().Where(notificationgroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.User.Delete().Where(user.EmailEQ("member-" + ns + "@example.com")).Exec(ctx)
	})

	notify := NewNotificationRepo(store, slog.Default())
	events := NewAlertEventRepo(store, slog.Default())
	smtp, err := notify.CreateChannel(ctx, bo.NotificationChannel{
		NamespaceUID: ns, Name: "smtp", Type: "SMTP", App: "OTHER",
		Endpoint: "127.0.0.1:1025", FromAddress: "from@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	emailTemplate, err := notify.CreateTemplate(ctx, bo.NotificationTemplate{
		NamespaceUID: ns, Name: "email", Type: "EMAIL", App: "OTHER", Subject: "alert", Body: "{{.summary}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	webhook, err := notify.CreateChannel(ctx, bo.NotificationChannel{
		NamespaceUID: ns, Name: "hook", Type: "WEBHOOK", App: "OTHER", Endpoint: "https://example.com/hook?token=secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	webhookTemplate, err := notify.CreateTemplate(ctx, bo.NotificationTemplate{
		NamespaceUID: ns, Name: "hook-tpl", Type: "WEBHOOK", App: "OTHER", Body: "{{.summary}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	userUID := "user-" + ns
	if _, err := store.client.User.Create().
		SetUID(userUID).
		SetEmail("member-" + ns + "@example.com").
		SetName("member").
		Save(ctx); err != nil {
		t.Fatal(err)
	}
	memberRow, err := store.client.Member.Create().
		SetUID("member-" + ns).
		SetNamespaceUID(ns).
		SetUserUID(userUID).
		SetStatus(member.StatusJOINED).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	group, err := notify.CreateGroup(ctx, bo.NotificationGroup{
		NamespaceUID: ns, Name: "ops", ChannelUID: smtp.UID, TemplateUID: emailTemplate.UID,
		Receivers: "manual@example.com", MemberUIDs: []string{memberRow.UID},
		WebhookChannelUIDs: []string{webhook.UID}, WebhookTemplateUID: webhookTemplate.UID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "prod", Labels: `{"env":"prod"}`,
		LevelUIDs: []string{"lv1"}, DatasourceUIDs: []string{"ds1"}, GroupUIDs: []string{group.UID},
	}); err != nil {
		t.Fatal(err)
	}
	rule := bo.AlertRule{UID: "rule-" + ns, NamespaceUID: ns, Expr: "up", Summary: "cpu"}
	slot := time.Now().UTC().Truncate(time.Minute)
	if err := events.ApplyReport(ctx, rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: "ds1", EvalSlot: slot,
		Hits: []bo.EvalHit{{Fingerprint: "fp1", Labels: `{"env":"prod"}`, Value: "1", LevelUID: "lv1", Summary: "fire"}},
	}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err := notify.ListTasks(ctx, ns, bo.NotificationTaskListFilter{}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("want 2 fanout tasks, got %d: %+v", len(tasks), tasks)
	}
	byChannel := map[string]bo.NotificationTask{}
	for _, task := range tasks {
		byChannel[task.ChannelUID] = task
	}
	emailTask, ok := byChannel[smtp.UID]
	if !ok {
		t.Fatal("missing smtp task")
	}
	if !strings.Contains(emailTask.Receivers, "manual@example.com") || !strings.Contains(emailTask.Receivers, "member-"+ns+"@example.com") {
		t.Fatalf("smtp receivers = %q", emailTask.Receivers)
	}
	if emailTask.Sender != "from@example.com" || emailTask.Subject != "alert" || emailTask.Body != "{{.summary}}" || emailTask.Content != "{{.summary}}" || emailTask.Status != "PENDING" {
		t.Fatalf("smtp task audit fields = %+v", emailTask)
	}
	webhookTask, ok := byChannel[webhook.UID]
	if !ok {
		t.Fatal("missing webhook task")
	}
	if webhookTask.Sender != "hook" || webhookTask.Receivers != "https://example.com/hook" || webhookTask.Body != "{{.summary}}" || webhookTask.Content != "{{.summary}}" || webhookTask.Status != "PENDING" {
		t.Fatalf("webhook task audit fields = %+v", webhookTask)
	}

	// 条件不匹配时不产生任务。
	if err := events.ApplyReport(ctx, rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: "ds2", EvalSlot: slot.Add(time.Minute),
		Hits: []bo.EvalHit{{Fingerprint: "fp2", Labels: `{"env":"dev"}`, Value: "1", LevelUID: "lv1", Summary: "dev"}},
	}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err = notify.ListTasks(ctx, ns, bo.NotificationTaskListFilter{}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("non-matching event created tasks: %d", len(tasks))
	}

	// 「我的告警」按实际通知任务的事件标识过滤：只有命中订阅并生成任务的
	// 事件才属于当前成员，不能把同规则下未投递给他的事件一起放行。
	eventUIDs, err := events.ListEventUIDsForUser(ctx, ns, memberRow.UID, []string{"member-" + ns + "@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if len(eventUIDs) != 1 {
		t.Fatalf("receiver event uids = %v, want 1", eventUIDs)
	}
	mine, total, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{
		View: "REALTIME", FilterByEvents: true, EventUIDs: eventUIDs,
	}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(mine) != 1 || mine[0].Summary != "fire" {
		t.Fatalf("my alerts = %+v total=%d, want only the matching event", mine, total)
	}
}
