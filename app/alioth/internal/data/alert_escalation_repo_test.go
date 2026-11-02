package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertescalationpolicy"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertescalationstep"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventaction"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventescalation"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

// newEscalationTestStore 连接集成测试库并应用 schema，与其它集成测试一致：
// 没有 ALIOTH_TEST_DSN 时跳过，绝不在开发库上执行。
func newEscalationTestStore(t *testing.T) (*Dependencies, context.Context) {
	t.Helper()
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
	return store, ctx
}

// cleanupEscalationNamespace 清掉本用例写入的升级数据与通知配置。
func cleanupEscalationNamespace(store *Dependencies, namespace string) func() {
	return func() {
		ctx := context.Background()
		_, _ = store.client.AlertEventEscalation.Delete().Where(alerteventescalation.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.AlertEventAction.Delete().Where(alerteventaction.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.AlertEscalationStep.Delete().Where(alertescalationstep.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.AlertEscalationPolicy.Delete().Where(alertescalationpolicy.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationGroup.Delete().Where(notificationgroup.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationChannel.Delete().Where(notificationchannel.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationTemplate.Delete().Where(notificationtemplate.NamespaceUIDEQ(namespace)).Exec(ctx)
	}
}

// escalationTestFixture 创建一条可直接投递的邮件通知组：渠道 + 模板 + 接收方。
// 返回通知组 UID 与渠道 UID，供断言任务生成使用。
func escalationTestFixture(ctx context.Context, t *testing.T, store *Dependencies, namespace string) (string, string) {
	t.Helper()
	channel, err := store.client.NotificationChannel.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetName("升级邮件渠道").
		SetType(notificationchannel.TypeSMTP).
		SetStatus(notificationchannel.StatusACTIVE).
		SetFromAddress("ops@example.com").
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	template, err := store.client.NotificationTemplate.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetName("升级邮件模板").
		SetType(notificationtemplate.TypeEMAIL).
		SetStatus(notificationtemplate.StatusACTIVE).
		SetSubject("[{{.phase}}] {{.summary}}").
		SetBody("{{.summary}}").
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	group, err := store.client.NotificationGroup.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespace).
		SetName("升级值班组").
		SetStatus(notificationgroup.StatusACTIVE).
		SetChannelUids([]string{channel.UID}).
		SetTemplateUids([]string{template.UID}).
		SetEmailReceivers("oncall@example.com").
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return group.UID, channel.UID
}

// escalationTestEvent 创建一条 20 分钟前开始、仍未认领的 FIRING 事件，
// 让 5 分钟和 15 分钟延迟的步骤在 now 时刻都已到期。
func escalationTestEvent(ctx context.Context, t *testing.T, store *Dependencies, namespace, uid, fingerprint string, now time.Time) {
	t.Helper()
	if err := store.client.AlertEvent.Create().
		SetUID(uid).
		SetNamespaceUID(namespace).
		SetRuleUID("rule-escalation").
		SetDatasourceUID("ds-escalation").
		SetFingerprint(fingerprint).
		SetStatus(alertevent.StatusFIRING).
		SetLabels(`{"env":"prod"}`).
		SetSummary("接口错误率过高").
		SetDescription("5xx 比例超过阈值").
		SetSampleValue("12.5").
		SetLevelUID("level-p1").
		SetStartsAt(now.Add(-20 * time.Minute)).
		SetEvalSlot(now).
		SetSuppressed(false).
		Exec(ctx); err != nil {
		t.Fatal(err)
	}
}

// TestAlertEscalationDispatch 覆盖升级调度的核心链路：
// 补齐 PENDING -> 到期生成通知任务并写 ESCALATED 时间线；
// 同事件同组同渠道已有 FIRING 任务时记 SKIPPED；
// 认领事件时未发送步骤立即取消。
func TestAlertEscalationDispatch(t *testing.T) {
	store, ctx := newEscalationTestStore(t)
	namespace := "ns-escalation-" + time.Now().Format("150405.000000")
	t.Cleanup(cleanupEscalationNamespace(store, namespace))
	repo := NewAlertEscalationRepo(store, slog.Default())
	groupUID, _ := escalationTestFixture(ctx, t, store, namespace)
	now := time.Now().UTC()

	newPolicy := func(name string) *bo.AlertEscalationPolicy {
		t.Helper()
		policy, err := repo.CreatePolicy(ctx, bo.AlertEscalationPolicy{
			NamespaceUID: namespace,
			Name:         name,
			Description:  "测试升级策略",
			Labels:       `{"env":"prod"}`,
			Steps: []bo.AlertEscalationStep{
				{DelayMinutes: 5, NotificationGroupUIDs: []string{groupUID}},
				{DelayMinutes: 15, NotificationGroupUIDs: []string{groupUID}},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		// 调度只升级策略创建之后开始的事件；这里把创建时间前移，模拟早已存在的策略。
		if _, err := store.db.ExecContext(ctx,
			`UPDATE alert_escalation_policies SET created_at = $1 WHERE uid = $2`,
			now.Add(-24*time.Hour), policy.UID); err != nil {
			t.Fatal(err)
		}
		return policy
	}

	eventUID := id.New()
	escalationTestEvent(ctx, t, store, namespace, eventUID, "fp-escalation-1", now)
	policy := newPolicy("生产告警升级")

	// 1) 补齐 PENDING 行：两步都应建立，重复执行不应重复建行。
	if _, err := repo.FillPendingEscalations(ctx, now.Add(-24*time.Hour), now); err != nil {
		t.Fatalf("fill pending: %v", err)
	}
	if _, err := repo.FillPendingEscalations(ctx, now.Add(-24*time.Hour), now); err != nil {
		t.Fatalf("fill pending second pass: %v", err)
	}
	rows, err := store.client.AlertEventEscalation.Query().Where(
		alerteventescalation.EventUIDEQ(eventUID),
		alerteventescalation.PolicyUIDEQ(policy.UID),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("escalation rows = %d, want 2", len(rows))
	}

	// 2) 到期发送：两步都已到期，各生成一条通知任务并置 SENT。
	if _, err := repo.CancelStaleEscalations(ctx, now); err != nil {
		t.Fatalf("cancel stale: %v", err)
	}
	if _, err := repo.SendDueEscalations(ctx, now); err != nil {
		t.Fatalf("send due: %v", err)
	}
	rows, err = store.client.AlertEventEscalation.Query().Where(
		alerteventescalation.EventUIDEQ(eventUID),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sentCount := 0
	for _, row := range rows {
		if row.Status == alerteventescalation.StatusSENT {
			sentCount++
			if row.SentAt == nil {
				t.Fatal("sent escalation row has no sent_at")
			}
		}
	}
	if sentCount != 1 {
		statuses := make([]string, 0, len(rows))
		for _, row := range rows {
			statuses = append(statuses, string(row.Status))
		}
		t.Fatalf("sent rows = %d, want 1 (statuses: %v)", sentCount, statuses)
	}
	tasks, err := store.client.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(eventUID),
		notificationtask.PhaseEQ(notificationtask.PhaseFIRING),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// 两步的通知组和渠道相同，唯一约束只允许一条任务，第二步应记 SKIPPED。
	if len(tasks) != 1 {
		t.Fatalf("notification tasks = %d, want 1", len(tasks))
	}
	skipped := 0
	for _, row := range rows {
		if row.StepOrder == 1 {
			if row.Status != alerteventescalation.StatusSKIPPED {
				t.Fatalf("second step status = %s, want SKIPPED", row.Status)
			}
			skipped++
		}
	}
	if skipped != 1 {
		t.Fatal("second step should be SKIPPED because the channel already has a FIRING task")
	}
	actions, err := store.client.AlertEventAction.Query().Where(
		alerteventaction.EventUIDEQ(eventUID),
		alerteventaction.ActionEQ("ESCALATED"),
	).Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if actions != 1 {
		t.Fatalf("ESCALATED actions = %d, want 1", actions)
	}

	// 3) 认领事件：未发送的步骤立即 CANCELLED，之后不再生成任务。
	ackEventUID := id.New()
	escalationTestEvent(ctx, t, store, namespace, ackEventUID, "fp-escalation-2", now)
	if _, err := repo.FillPendingEscalations(ctx, now.Add(-24*time.Hour), now); err != nil {
		t.Fatalf("fill pending for ack case: %v", err)
	}
	acked, err := NewAlertEventRepo(store, slog.Default()).AckEvent(ctx, namespace, ackEventUID, "user-ack")
	if err != nil {
		t.Fatalf("ack event: %v", err)
	}
	if acked.AckedBy != "user-ack" {
		t.Fatalf("acked_by = %q", acked.AckedBy)
	}
	pending, err := store.client.AlertEventEscalation.Query().Where(
		alerteventescalation.EventUIDEQ(ackEventUID),
		alerteventescalation.StatusEQ(alerteventescalation.StatusPENDING),
	).Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pending != 0 {
		t.Fatalf("pending rows after ack = %d, want 0", pending)
	}
	if _, err := repo.SendDueEscalations(ctx, now); err != nil {
		t.Fatalf("send due after ack: %v", err)
	}
	ackTasks, err := store.client.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(ackEventUID),
	).Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ackTasks != 0 {
		t.Fatalf("tasks after ack = %d, want 0", ackTasks)
	}
}
