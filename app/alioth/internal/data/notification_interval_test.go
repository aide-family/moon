package data

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestNotificationFilterDatasourceLevel 校验订阅和 OnCall 的数据源级别过滤：
// 条件之间是 AND，空条件不限制，事件缺少数据源级别时不匹配非空条件。
func TestNotificationFilterDatasourceLevel(t *testing.T) {
	event := notifyEvent{
		LevelUID: "lv1", DatasourceUID: "ds1", DatasourceLevelUID: "dl1",
		Labels: `{"env":"prod"}`,
	}
	subscriptionTests := []struct {
		name   string
		filter *ent.NotificationSubscription
		event  notifyEvent
		want   bool
	}{
		{
			name:   "empty datasource level matches any event",
			filter: &ent.NotificationSubscription{Labels: "{}"},
			event:  event,
			want:   true,
		},
		{
			name:   "datasource level match",
			filter: &ent.NotificationSubscription{Labels: "{}", DatasourceLevelUids: []string{"dl1", "dl2"}},
			event:  event,
			want:   true,
		},
		{
			name:   "datasource level mismatch",
			filter: &ent.NotificationSubscription{Labels: "{}", DatasourceLevelUids: []string{"dl2"}},
			event:  event,
			want:   false,
		},
		{
			name: "all filters must match",
			filter: &ent.NotificationSubscription{
				Labels: `{"env":"prod"}`, LevelUids: []string{"lv1"},
				DatasourceUids: []string{"ds1"}, DatasourceLevelUids: []string{"dl1"},
			},
			event: event,
			want:  true,
		},
		{
			name: "datasource level mismatch breaks AND",
			filter: &ent.NotificationSubscription{
				Labels: `{"env":"prod"}`, LevelUids: []string{"lv1"},
				DatasourceUids: []string{"ds1"}, DatasourceLevelUids: []string{"dl2"},
			},
			event: event,
			want:  false,
		},
		{
			name:   "missing event datasource level does not match",
			filter: &ent.NotificationSubscription{Labels: "{}", DatasourceLevelUids: []string{"dl1"}},
			event:  notifyEvent{LevelUID: "lv1", DatasourceUID: "ds1", Labels: "{}"},
			want:   false,
		},
	}
	for _, item := range subscriptionTests {
		t.Run("subscription/"+item.name, func(t *testing.T) {
			if got := matchSubscription(item.filter, item.event); got != item.want {
				t.Fatalf("matchSubscription = %v, want %v", got, item.want)
			}
		})
	}
	onCallTests := []struct {
		name  string
		event notifyEvent
		want  bool
	}{
		{
			name:  "datasource level match",
			event: event,
			want:  true,
		},
		{
			name:  "datasource level mismatch",
			event: notifyEvent{LevelUID: "lv1", DatasourceUID: "ds1", DatasourceLevelUID: "dl2", Labels: `{"env":"prod"}`},
			want:  false,
		},
	}
	for _, item := range onCallTests {
		t.Run("oncall/"+item.name, func(t *testing.T) {
			onCall := &ent.NotificationOnCall{
				Labels: `{"env":"prod"}`, LevelUids: []string{"lv1"},
				DatasourceUids: []string{"ds1"}, DatasourceLevelUids: []string{"dl1"},
			}
			if got := matchOnCall(onCall, item.event); got != item.want {
				t.Fatalf("matchOnCall = %v, want %v", got, item.want)
			}
		})
	}
}

// newNotificationTestChannel 创建一条投递渠道和配套模板。
func newNotificationTestChannel(t *testing.T, ctx context.Context, notify repository.NotificationRepo, ns, name, kind string) (*bo.NotificationChannel, *bo.NotificationTemplate) {
	t.Helper()
	channel := bo.NotificationChannel{
		NamespaceUID: ns, Name: name + "-channel", Type: kind, App: "OTHER",
		Endpoint: "127.0.0.1:1025", FromAddress: "from@example.com",
	}
	if kind == "WEBHOOK" {
		channel.Endpoint = "https://example.com/" + name + "?token=secret"
	}
	created, err := notify.CreateChannel(ctx, channel)
	if err != nil {
		t.Fatal(err)
	}
	templateKind, subject := "EMAIL", "alert"
	switch kind {
	case "WEBHOOK":
		templateKind, subject = "WEBHOOK", ""
	case "SMS":
		templateKind = "SMS_ALICLOUD"
	}
	template, err := notify.CreateTemplate(ctx, bo.NotificationTemplate{
		NamespaceUID: ns, Name: name + "-template", Type: templateKind, App: "OTHER", Subject: subject, Body: "{{.summary}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	return created, template
}

// newNotificationTestEmailGroup 创建只发邮件的通知组。
func newNotificationTestEmailGroup(t *testing.T, ctx context.Context, notify repository.NotificationRepo, ns, name, receivers string) *bo.NotificationGroup {
	t.Helper()
	channel, template := newNotificationTestChannel(t, ctx, notify, ns, name, "SMTP")
	group, err := notify.CreateGroup(ctx, bo.NotificationGroup{
		NamespaceUID: ns, Name: name + "-group", ChannelUIDs: []string{channel.UID},
		TemplateUIDs: []string{template.UID}, EmailReceivers: receivers,
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

// newNotificationTestSMSGroup 创建只发短信的通知组。
func newNotificationTestSMSGroup(t *testing.T, ctx context.Context, notify repository.NotificationRepo, ns, name, receivers string) *bo.NotificationGroup {
	t.Helper()
	channel, template := newNotificationTestChannel(t, ctx, notify, ns, name, "SMS")
	group, err := notify.CreateGroup(ctx, bo.NotificationGroup{
		NamespaceUID: ns, Name: name + "-group", ChannelUIDs: []string{channel.UID},
		TemplateUIDs: []string{template.UID}, SMSReceivers: receivers,
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

// TestNormalizeContact 校验联系人归一化：邮箱小写、手机号去格式字符，
// 带国家码和不带国家码的同一号码在去重时能识别为同一联系人。
func TestNormalizeContact(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: " Shared@Example.com ", want: "shared@example.com"},
		{in: "+86 138-0000-0000", want: "8613800000000"},
		{in: "(138) 0000.0000", want: "13800000000"},
		{in: "not-a-contact", want: "not-a-contact"},
		{in: "", want: ""},
	}
	for _, item := range cases {
		if got := normalizeContact(item.in); got != item.want {
			t.Fatalf("normalizeContact(%q) = %q, want %q", item.in, got, item.want)
		}
	}
	tracker := &dispatchTracker{contacts: map[string]struct{}{}, scheduledChannels: map[string]struct{}{}}
	tracker.contacts[normalizeContact("+86 138-0000-0000")] = struct{}{}
	if !tracker.contactDelivered("13800000000") {
		t.Fatal("phone with country code was not deduplicated")
	}
	if tracker.contactDelivered("13900000000") {
		t.Fatal("different phone was treated as delivered")
	}
}

// registerReportCleanup 清理本测试用唯一数据源标识写入的评测批次。
// 每个测试用带命名空间后缀的数据源标识，避免并发跑测试时互相删除批次。
func registerReportCleanup(t *testing.T, ctx context.Context, store *Dependencies, datasourceUID string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.DatasourceUIDEQ(datasourceUID)).Exec(ctx)
	})
}

// reportRule 返回测试用的规则标识。规则标识保持短小（历史库 rule_uid 为 VARCHAR(64)），
// 唯一性由每个测试独立的 datasourceUID 保证。
func reportRule(fingerprint, namespaceUID string) bo.AlertRule {
	return bo.AlertRule{UID: "rule-" + fingerprint, NamespaceUID: namespaceUID, Expr: "up", Summary: "cpu"}
}

// applyFiringReport 写入一条 FIRING 上报。
func applyFiringReport(t *testing.T, ctx context.Context, events repository.AlertEventRepo, ns, datasourceUID, fingerprint string, slot time.Time) {
	t.Helper()
	rule := reportRule(fingerprint, ns)
	if err := events.ApplyReport(ctx, rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: datasourceUID, EvalSlot: slot,
		Hits: []bo.EvalHit{{Fingerprint: fingerprint, Labels: "{}", Value: "1", LevelUID: "lv1", Summary: "fire"}},
	}); err != nil {
		t.Fatal(err)
	}
}

// applyRecoveryReport 写入一条没有命中的上报，触发恢复通知。
func applyRecoveryReport(t *testing.T, ctx context.Context, events repository.AlertEventRepo, ns, datasourceUID, fingerprint string, slot time.Time) {
	t.Helper()
	rule := reportRule(fingerprint, ns)
	if err := events.ApplyReport(ctx, rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: datasourceUID, EvalSlot: slot, Hits: nil,
	}); err != nil {
		t.Fatal(err)
	}
}

// tasksByFingerprint 读取同一指纹的全部任务，按写入顺序返回。
func tasksByFingerprint(t *testing.T, ctx context.Context, store *Dependencies, ns, fingerprint string) []*ent.NotificationTask {
	t.Helper()
	rows, err := store.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(ns),
		notificationtask.FingerprintEQ(fingerprint),
	).Order(ent.Asc(notificationtask.FieldID)).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

// countTaskPhase 统计指定阶段的任务数。
func countTaskPhase(tasks []*ent.NotificationTask, phase notificationtask.Phase) int {
	count := 0
	for _, task := range tasks {
		if task.Phase == phase {
			count++
		}
	}
	return count
}

// backdateFiringTasks 把同一指纹的 FIRING 任务创建时间改到 age 之前，模拟历史通知。
func backdateFiringTasks(t *testing.T, ctx context.Context, store *Dependencies, ns, fingerprint string, age time.Duration) {
	t.Helper()
	if _, err := store.db.ExecContext(ctx,
		`UPDATE notification_tasks SET created_at = now() - ($1 * interval '1 second')
		 WHERE namespace_uid = $2 AND fingerprint = $3 AND phase = 'FIRING'`,
		int64(age.Seconds()), ns, fingerprint,
	); err != nil {
		t.Fatal(err)
	}
}

// TestNotificationDispatchRecipientDedupe 校验同一事件里一个联系人、一个 Webhook 渠道最多安排一次：
// 两个通知组、订阅 + OnCall 两套路由同时命中时都只保留先处理的一条任务。
func TestNotificationDispatchRecipientDedupe(t *testing.T) {
	t.Run("across groups", func(t *testing.T) {
		ctx, store, notify, events, ns := onCallTestEnvironment(t)
		datasourceUID := "ds-" + ns
		registerReportCleanup(t, ctx, store, datasourceUID)
		slot := time.Now().UTC().Truncate(time.Minute)

		// 两个通知组用不同 SMTP 渠道投递同一个邮箱：只保留一条任务。
		groupA := newNotificationTestEmailGroup(t, ctx, notify, ns, "shared-a", "shared@example.com")
		groupB := newNotificationTestEmailGroup(t, ctx, notify, ns, "shared-b", "shared@example.com")
		if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
			NamespaceUID: ns, Name: "shared-sub", Labels: "{}",
			GroupUIDs: []string{groupA.UID, groupB.UID},
		}); err != nil {
			t.Fatal(err)
		}
		applyFiringReport(t, ctx, events, ns, datasourceUID, "dedupe-groups", slot)
		tasks := tasksByFingerprint(t, ctx, store, ns, "dedupe-groups")
		if len(tasks) != 1 || tasks[0].Receivers != "shared@example.com" {
			t.Fatalf("group dedupe tasks = %+v, want 1 task keeping shared@example.com", tasks)
		}
	})

	t.Run("across subscription and oncall", func(t *testing.T) {
		ctx, store, notify, events, ns := onCallTestEnvironment(t)
		datasourceUID := "ds-" + ns
		registerReportCleanup(t, ctx, store, datasourceUID)
		slot := time.Now().UTC().Truncate(time.Minute)

		// 订阅命中的通知组排在前：订阅 + OnCall 两套路由的同一邮箱只发订阅这一条。
		subscriptionGroup := newNotificationTestEmailGroup(t, ctx, notify, ns, "route-sub", "route@example.com")
		onCallGroup := newNotificationTestEmailGroup(t, ctx, notify, ns, "route-oncall", "route@example.com")
		subscription, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
			NamespaceUID: ns, Name: "route-sub", Labels: "{}", GroupUIDs: []string{subscriptionGroup.UID},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{
			NamespaceUID: ns, Name: "route-oncall", Labels: "{}",
			Groups: []bo.NotificationOnCallGroup{{GroupUID: onCallGroup.UID}},
		}); err != nil {
			t.Fatal(err)
		}
		applyFiringReport(t, ctx, events, ns, datasourceUID, "dedupe-routes", slot)
		tasks := tasksByFingerprint(t, ctx, store, ns, "dedupe-routes")
		if len(tasks) != 1 {
			t.Fatalf("route dedupe tasks = %d, want 1: %+v", len(tasks), tasks)
		}
		if tasks[0].GroupUID != subscriptionGroup.UID || tasks[0].SubscriptionUID != subscription.UID {
			t.Fatalf("deduped task = group %q subscription %q, want subscription route group %q",
				tasks[0].GroupUID, tasks[0].SubscriptionUID, subscriptionGroup.UID)
		}
	})

	t.Run("across webhook groups", func(t *testing.T) {
		ctx, store, notify, events, ns := onCallTestEnvironment(t)
		datasourceUID := "ds-" + ns
		registerReportCleanup(t, ctx, store, datasourceUID)
		slot := time.Now().UTC().Truncate(time.Minute)

		// 两个通知组绑定同一个 Webhook 渠道：只安排一条 Webhook 任务。
		hookChannel, hookTemplate := newNotificationTestChannel(t, ctx, notify, ns, "hook-shared", "WEBHOOK")
		hookGroupA, err := notify.CreateGroup(ctx, bo.NotificationGroup{
			NamespaceUID: ns, Name: "hook-a-group",
			ChannelUIDs: []string{hookChannel.UID}, TemplateUIDs: []string{hookTemplate.UID},
		})
		if err != nil {
			t.Fatal(err)
		}
		hookGroupB, err := notify.CreateGroup(ctx, bo.NotificationGroup{
			NamespaceUID: ns, Name: "hook-b-group",
			ChannelUIDs: []string{hookChannel.UID}, TemplateUIDs: []string{hookTemplate.UID},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
			NamespaceUID: ns, Name: "hook-sub", Labels: "{}",
			GroupUIDs: []string{hookGroupA.UID, hookGroupB.UID},
		}); err != nil {
			t.Fatal(err)
		}
		applyFiringReport(t, ctx, events, ns, datasourceUID, "dedupe-webhook", slot)
		tasks := tasksByFingerprint(t, ctx, store, ns, "dedupe-webhook")
		if len(tasks) != 1 || tasks[0].ChannelUID != hookChannel.UID {
			t.Fatalf("webhook dedupe tasks = %+v, want 1 task on channel %q", tasks, hookChannel.UID)
		}
	})

	t.Run("across phone formats", func(t *testing.T) {
		ctx, store, notify, events, ns := onCallTestEnvironment(t)
		datasourceUID := "ds-" + ns
		registerReportCleanup(t, ctx, store, datasourceUID)
		slot := time.Now().UTC().Truncate(time.Minute)

		// 同一个手机号带不带国家码、夹杂格式字符时只发一条短信。
		groupA := newNotificationTestSMSGroup(t, ctx, notify, ns, "sms-a", "+86 138-0000-0000")
		groupB := newNotificationTestSMSGroup(t, ctx, notify, ns, "sms-b", "13800000000")
		if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
			NamespaceUID: ns, Name: "sms-sub", Labels: "{}",
			GroupUIDs: []string{groupA.UID, groupB.UID},
		}); err != nil {
			t.Fatal(err)
		}
		applyFiringReport(t, ctx, events, ns, datasourceUID, "dedupe-phone", slot)
		tasks := tasksByFingerprint(t, ctx, store, ns, "dedupe-phone")
		if len(tasks) != 1 {
			t.Fatalf("phone dedupe tasks = %d, want 1: %+v", len(tasks), tasks)
		}
	})
}

// TestNotificationRepeatInterval 校验同一指纹的 FIRING 重复通知间隔：
// 间隔内抑制、间隔外放行、不同指纹互不影响，恢复通知不受间隔限制。
func TestNotificationRepeatInterval(t *testing.T) {
	ctx, store, notify, events, ns := onCallTestEnvironment(t)
	t.Cleanup(func() { _ = store.history.dropTenant(context.Background(), ns) })
	datasourceUID := "ds-" + ns
	registerReportCleanup(t, ctx, store, datasourceUID)
	group := newNotificationTestEmailGroup(t, ctx, notify, ns, "repeat", "repeat@example.com")
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "repeat-sub", Labels: "{}", RepeatIntervalSeconds: 3600,
		GroupUIDs: []string{group.UID},
	}); err != nil {
		t.Fatal(err)
	}
	slot := time.Now().UTC().Truncate(time.Minute)

	// 第一次 FIRING：正常生成任务。
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-repeat", slot)
	tasks := tasksByFingerprint(t, ctx, store, ns, "fp-repeat")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 1 {
		t.Fatalf("initial firing tasks = %+v, want 1", tasks)
	}

	// 30 分钟前发过 FIRING 通知：恢复仍然照常发送，不受重复间隔限制。
	backdateFiringTasks(t, ctx, store, ns, "fp-repeat", 30*time.Minute)
	applyRecoveryReport(t, ctx, events, ns, datasourceUID, "fp-repeat", slot.Add(time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-repeat")
	if countTaskPhase(tasks, notificationtask.PhaseRECOVERED) != 1 {
		t.Fatalf("recovery within interval tasks = %+v, want 1 recovery", tasks)
	}

	// 再次 FIRING 落在 1 小时间隔内：整条事件不再生成任务。
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-repeat", slot.Add(2*time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-repeat")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 1 {
		t.Fatalf("firing within interval tasks = %+v, want still 1 firing", tasks)
	}

	// 把历史 FIRING 挪到 2 小时前，间隔外允许再次通知。
	backdateFiringTasks(t, ctx, store, ns, "fp-repeat", 2*time.Hour)
	applyRecoveryReport(t, ctx, events, ns, datasourceUID, "fp-repeat", slot.Add(3*time.Minute))
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-repeat", slot.Add(4*time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-repeat")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 2 {
		t.Fatalf("firing after interval tasks = %+v, want 2 firing", tasks)
	}

	// 不同指纹不受同一告警的历史通知影响。
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-other", slot.Add(5*time.Minute))
	otherTasks := tasksByFingerprint(t, ctx, store, ns, "fp-other")
	if countTaskPhase(otherTasks, notificationtask.PhaseFIRING) != 1 {
		t.Fatalf("different fingerprint tasks = %+v, want 1 firing", otherTasks)
	}
}

// TestNotificationDispatchMaxInterval 校验命中的订阅和 OnCall 间隔不同时取最大值：
// 订阅 7200 秒、OnCall 600 秒时，1.5 小时前的通知仍会抑制 FIRING。
func TestNotificationDispatchMaxInterval(t *testing.T) {
	ctx, store, notify, events, ns := onCallTestEnvironment(t)
	t.Cleanup(func() { _ = store.history.dropTenant(context.Background(), ns) })
	datasourceUID := "ds-" + ns
	registerReportCleanup(t, ctx, store, datasourceUID)
	groupA := newNotificationTestEmailGroup(t, ctx, notify, ns, "max-sub", "max-sub@example.com")
	groupB := newNotificationTestEmailGroup(t, ctx, notify, ns, "max-oncall", "max-oncall@example.com")
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "max-sub", Labels: "{}", RepeatIntervalSeconds: 7200,
		GroupUIDs: []string{groupA.UID},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: ns, Name: "max-oncall", Labels: "{}", RepeatIntervalSeconds: 600,
		Groups: []bo.NotificationOnCallGroup{{GroupUID: groupB.UID}},
	}); err != nil {
		t.Fatal(err)
	}
	slot := time.Now().UTC().Truncate(time.Minute)

	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-max", slot)
	tasks := tasksByFingerprint(t, ctx, store, ns, "fp-max")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 2 {
		t.Fatalf("initial max interval tasks = %+v, want 2 firing", tasks)
	}

	// 90 分钟前发过通知：超过 OnCall 的 600 秒但仍在订阅的 7200 秒内，应被抑制。
	backdateFiringTasks(t, ctx, store, ns, "fp-max", 90*time.Minute)
	applyRecoveryReport(t, ctx, events, ns, datasourceUID, "fp-max", slot.Add(time.Minute))
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-max", slot.Add(2*time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-max")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 2 {
		t.Fatalf("firing within max interval tasks = %+v, want still 2 firing", tasks)
	}

	// 超过 7200 秒后允许再次通知。
	backdateFiringTasks(t, ctx, store, ns, "fp-max", 3*time.Hour)
	applyRecoveryReport(t, ctx, events, ns, datasourceUID, "fp-max", slot.Add(3*time.Minute))
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-max", slot.Add(4*time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-max")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 4 {
		t.Fatalf("firing after max interval tasks = %+v, want 4 firing", tasks)
	}
}

// TestNotificationRecoveryRequiresFiring 校验恢复只在同一指纹存在未取消的 FIRING 任务时发送：
// 从未发过的指纹和被取消的 FIRING 都不补发恢复通知。
func TestNotificationRecoveryRequiresFiring(t *testing.T) {
	ctx, store, notify, events, ns := onCallTestEnvironment(t)
	t.Cleanup(func() { _ = store.history.dropTenant(context.Background(), ns) })
	datasourceUID := "ds-" + ns
	registerReportCleanup(t, ctx, store, datasourceUID)
	group := newNotificationTestEmailGroup(t, ctx, notify, ns, "recovery", "recovery@example.com")
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "recovery-sub", Labels: "{}", GroupUIDs: []string{group.UID},
	}); err != nil {
		t.Fatal(err)
	}

	// 直接分发一条从未有 FIRING 任务的恢复事件。
	tx, err := store.client.Tx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatchEventNotifications(ctx, tx, notifyEvent{
		NamespaceUID: ns, EventUID: "evt-never-" + ns, RuleUID: "rule-never-" + ns,
		DatasourceUID: datasourceUID, LevelUID: "lv1", Phase: "RECOVERED",
		Summary: "never", Labels: "{}", Fingerprint: "fp-never",
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	rows, err := store.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(ns),
		notificationtask.EventUIDEQ("evt-never-"+ns),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("recovery without prior firing created %d tasks", len(rows))
	}

	// 升级前的 FIRING 任务没有 fingerprint，恢复仍按事件标识兜底发送。
	legacyEventUID := "evt-legacy-" + ns
	if err := store.client.NotificationTask.Create().
		SetUID("task-legacy-" + ns).
		SetNamespaceUID(ns).
		SetEventUID(legacyEventUID).
		SetPhase(notificationtask.PhaseFIRING).
		SetStatus(notificationtask.StatusPENDING).
		SetChannelUID("legacy-channel").
		Exec(ctx); err != nil {
		t.Fatal(err)
	}
	legacyTx, err := store.client.Tx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatchEventNotifications(ctx, legacyTx, notifyEvent{
		NamespaceUID: ns, EventUID: legacyEventUID, RuleUID: "rule-legacy-" + ns,
		DatasourceUID: datasourceUID, LevelUID: "lv1", Phase: "RECOVERED",
		Summary: "legacy", Labels: "{}", Fingerprint: "fp-legacy",
	}); err != nil {
		t.Fatal(err)
	}
	if err := legacyTx.Commit(); err != nil {
		t.Fatal(err)
	}
	legacyTasks, err := store.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(ns),
		notificationtask.EventUIDEQ(legacyEventUID),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacyTasks) != 2 || countTaskPhase(legacyTasks, notificationtask.PhaseRECOVERED) != 1 {
		t.Fatalf("legacy firing recovery tasks = %+v, want 2 tasks including 1 recovery", legacyTasks)
	}

	// FIRING 任务被取消后，恢复同样跳过。
	slot := time.Now().UTC().Truncate(time.Minute)
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-cancelled", slot)
	tasks := tasksByFingerprint(t, ctx, store, ns, "fp-cancelled")
	if len(tasks) != 1 {
		t.Fatalf("cancelled firing tasks = %+v, want 1", tasks)
	}
	if _, err := store.client.NotificationTask.Update().Where(
		notificationtask.UIDEQ(tasks[0].UID),
	).SetStatus(notificationtask.StatusCANCELLED).Save(ctx); err != nil {
		t.Fatal(err)
	}
	applyRecoveryReport(t, ctx, events, ns, datasourceUID, "fp-cancelled", slot.Add(time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-cancelled")
	if countTaskPhase(tasks, notificationtask.PhaseRECOVERED) != 0 {
		t.Fatalf("recovery for cancelled firing tasks = %+v, want no recovery", tasks)
	}

	// 彻底失败的 FIRING 没有送达用户，恢复也不补发。
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-failed", slot.Add(2*time.Minute))
	failedTasks := tasksByFingerprint(t, ctx, store, ns, "fp-failed")
	if len(failedTasks) != 1 {
		t.Fatalf("failed firing tasks = %+v, want 1", failedTasks)
	}
	if _, err := store.client.NotificationTask.Update().Where(
		notificationtask.UIDEQ(failedTasks[0].UID),
	).SetStatus(notificationtask.StatusFAILED).Save(ctx); err != nil {
		t.Fatal(err)
	}
	applyRecoveryReport(t, ctx, events, ns, datasourceUID, "fp-failed", slot.Add(3*time.Minute))
	failedTasks = tasksByFingerprint(t, ctx, store, ns, "fp-failed")
	if countTaskPhase(failedTasks, notificationtask.PhaseRECOVERED) != 0 {
		t.Fatalf("recovery for failed firing tasks = %+v, want no recovery", failedTasks)
	}
}

// TestNotificationDatasourceLevelDispatch 校验 FIRING 分发把数据源自身的级别带进过滤条件。
func TestNotificationDatasourceLevelDispatch(t *testing.T) {
	ctx, store, notify, events, ns := onCallTestEnvironment(t)
	datasourceUID := "ds-" + ns
	registerReportCleanup(t, ctx, store, datasourceUID)
	levelUID := "ds-level-" + ns
	if err := store.client.Datasource.Create().
		SetUID(datasourceUID).
		SetNamespaceUID(ns).
		SetName("prom").
		SetType("METRICS").
		SetDriver("METRICS_PROMETHEUS").
		SetEndpoint("http://127.0.0.1:9090").
		SetLevelUID(levelUID).
		Exec(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = store.client.Datasource.Delete().Where(datasource.UIDEQ(datasourceUID)).Exec(ctx)
	})
	group := newNotificationTestEmailGroup(t, ctx, notify, ns, "ds-level", "ds-level@example.com")
	subscription, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "ds-level-match", Labels: "{}",
		DatasourceLevelUIDs: []string{levelUID}, GroupUIDs: []string{group.UID},
	})
	if err != nil {
		t.Fatal(err)
	}
	slot := time.Now().UTC().Truncate(time.Minute)

	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-ds-level", slot)
	tasks := tasksByFingerprint(t, ctx, store, ns, "fp-ds-level")
	if countTaskPhase(tasks, notificationtask.PhaseFIRING) != 1 {
		t.Fatalf("datasource level match tasks = %+v, want 1 firing", tasks)
	}

	// 数据源级别不匹配的订阅不产生任务。
	if _, err := notify.UpdateSubscriptionStatus(ctx, ns, subscription.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "ds-level-mismatch", Labels: "{}",
		DatasourceLevelUIDs: []string{"other-" + ns}, GroupUIDs: []string{group.UID},
	}); err != nil {
		t.Fatal(err)
	}
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-ds-mismatch", slot.Add(time.Minute))
	tasks = tasksByFingerprint(t, ctx, store, ns, "fp-ds-mismatch")
	if len(tasks) != 0 {
		t.Fatalf("datasource level mismatch created tasks: %+v", tasks)
	}
}
