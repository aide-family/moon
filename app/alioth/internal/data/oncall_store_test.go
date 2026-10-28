package data

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncall"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncallgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// onCallTestEnvironment 打开集成测试库，注册按命名空间清理，返回租户隔离的测试标识。
func onCallTestEnvironment(t *testing.T) (context.Context, *Dependencies, repository.NotificationRepo, repository.AlertEventRepo, string) {
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
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	ns := "ns-oncall-" + strings.ReplaceAll(t.Name(), "/", "-") + "-" + time.Now().Format("150405.000")
	t.Cleanup(func() {
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationOnCallGroup.Delete().Where(notificationoncallgroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationOnCall.Delete().Where(notificationoncall.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationSubscriptionGroup.Delete().Where(notificationsubscriptiongroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationSubscription.Delete().Where(notificationsubscription.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEventFiring.Delete().Where(alerteventfiring.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.RuleUIDContains(ns)).Exec(ctx)
		_, _ = store.client.NotificationGroup.Delete().Where(notificationgroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationTemplate.Delete().Where(notificationtemplate.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.NotificationChannel.Delete().Where(notificationchannel.NamespaceUIDEQ(ns)).Exec(ctx)
	})
	return ctx, store, NewNotificationRepo(store, slog.Default()), NewAlertEventRepo(store, slog.Default()), ns
}

// newOnCallTestGroup 创建一套邮件渠道、模板和通知组。
func newOnCallTestGroup(t *testing.T, ctx context.Context, notify repository.NotificationRepo, ns, name string) *bo.NotificationGroup {
	t.Helper()
	channel, err := notify.CreateChannel(ctx, bo.NotificationChannel{
		NamespaceUID: ns, Name: name + "-smtp", Type: "SMTP", App: "OTHER",
		Endpoint: "127.0.0.1:1025", FromAddress: "from@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	template, err := notify.CreateTemplate(ctx, bo.NotificationTemplate{
		NamespaceUID: ns, Name: name + "-tpl", Type: "EMAIL", App: "OTHER", Subject: "alert", Body: "{{.summary}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := notify.CreateGroup(ctx, bo.NotificationGroup{
		NamespaceUID: ns, Name: name + "-group",
		ChannelUIDs: []string{channel.UID}, TemplateUIDs: []string{template.UID},
		EmailReceivers: name + "@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	return group
}

func allWeekdays() []string {
	return []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}
}

// onCallWindowAround 生成以 at 为基准、从 offsetHours 开始、持续两小时的窗口，
// 星期覆盖每一天，这样测试结果不受当前星期影响。
func onCallWindowAround(at time.Time, offsetHours int) []bo.NotificationOnCallWindow {
	start := at.Add(time.Duration(offsetHours) * time.Hour)
	end := start.Add(2 * time.Hour)
	return []bo.NotificationOnCallWindow{{
		Weekdays: allWeekdays(), Start: start.Format("15:04"), End: end.Format("15:04"),
	}}
}

func TestOnCallRepoLifecycle(t *testing.T) {
	ctx, store, notify, _, ns := onCallTestEnvironment(t)
	groupA := newOnCallTestGroup(t, ctx, notify, ns, "a")
	groupB := newOnCallTestGroup(t, ctx, notify, ns, "b")

	created, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: ns, Name: "primary", Labels: `{"env":"prod"}`,
		LevelUIDs: []string{"lv1"}, DatasourceUIDs: []string{"ds1"},
		Groups: []bo.NotificationOnCallGroup{
			{GroupUID: groupA.UID, Windows: []bo.NotificationOnCallWindow{
				{Weekdays: []string{"friday", "MONDAY", "Monday"}, Start: "9:00", End: "18:00"},
			}},
			// 同一通知组重复出现时合并为一条绑定。
			{GroupUID: groupA.UID, Windows: []bo.NotificationOnCallWindow{
				{Weekdays: []string{"SATURDAY"}, Start: "10:00", End: "11:00"},
			}},
			{GroupUID: groupB.UID},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.UID == "" || created.Status != "ACTIVE" {
		t.Fatalf("created on-call = %+v", created)
	}
	if created.Timezone != "Asia/Shanghai" {
		t.Fatalf("default timezone = %q, want Asia/Shanghai", created.Timezone)
	}
	if len(created.Groups) != 2 {
		t.Fatalf("created groups = %+v, want 2", created.Groups)
	}
	groupsByUID := map[string]bo.NotificationOnCallGroup{}
	for _, group := range created.Groups {
		groupsByUID[group.GroupUID] = group
	}
	boundA, ok := groupsByUID[groupA.UID]
	if !ok || boundA.GroupName != groupA.Name {
		t.Fatalf("group A binding = %+v", boundA)
	}
	windows := boundA.Windows
	if len(windows) != 1 || windows[0].Start != "09:00" || windows[0].End != "18:00" || len(windows[0].Weekdays) != 2 || windows[0].Weekdays[0] != "MONDAY" || windows[0].Weekdays[1] != "FRIDAY" {
		t.Fatalf("normalized windows = %+v", windows)
	}
	if boundB, ok := groupsByUID[groupB.UID]; !ok || len(boundB.Windows) != 0 {
		t.Fatalf("group B binding = %+v", boundB)
	}
	links, err := store.client.NotificationOnCallGroup.Query().Where(
		notificationoncallgroup.NamespaceUIDEQ(ns),
		notificationoncallgroup.OncallUIDEQ(created.UID),
		notificationoncallgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("duplicate group bindings written: %d", len(links))
	}

	// 通知组绑定了 OnCall 时不能直接删除。
	if err := notify.DeleteGroup(ctx, ns, groupA.UID); !merr.IsConflict(err) {
		t.Fatalf("delete bound group error = %v, want conflict", err)
	}

	// 更新时替换绑定：旧的 groupA 绑定软删除，groupB 生效，时区改成规范名称。
	updated, err := notify.UpdateOnCall(ctx, bo.NotificationOnCall{
		UID: created.UID, NamespaceUID: ns, Name: "primary-renamed",
		Labels: `{"env":"staging"}`, LevelUIDs: []string{"lv2"}, DatasourceUIDs: nil,
		Timezone: "UTC",
		Groups: []bo.NotificationOnCallGroup{
			{GroupUID: groupB.UID},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "primary-renamed" || updated.Timezone != "UTC" || len(updated.LevelUIDs) != 1 || updated.LevelUIDs[0] != "lv2" {
		t.Fatalf("updated on-call = %+v", updated)
	}
	if len(updated.Groups) != 1 || updated.Groups[0].GroupUID != groupB.UID || updated.Groups[0].GroupName != groupB.Name {
		t.Fatalf("updated groups = %+v", updated.Groups)
	}
	if len(updated.Groups[0].Windows) != 0 {
		t.Fatalf("empty windows must mean all-day, got %+v", updated.Groups[0].Windows)
	}
	oldLinks, err := store.client.NotificationOnCallGroup.Query().Where(
		notificationoncallgroup.NamespaceUIDEQ(ns),
		notificationoncallgroup.OncallUIDEQ(created.UID),
		notificationoncallgroup.GroupUIDEQ(groupA.UID),
	).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, link := range oldLinks {
		if link.DeletedAt == nil {
			t.Fatal("replaced group binding is still active")
		}
	}

	items, total, err := notify.ListOnCalls(ctx, ns, 1, 10, "renamed")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].UID != created.UID {
		t.Fatalf("list on-calls = %+v total=%d", items, total)
	}
	if len(items[0].Groups) != 1 || items[0].Groups[0].GroupName != groupB.Name {
		t.Fatalf("list group names = %+v", items[0].Groups)
	}

	disabled, err := notify.UpdateOnCallStatus(ctx, ns, created.UID, "DISABLED")
	if err != nil {
		t.Fatal(err)
	}
	if disabled.Status != "DISABLED" {
		t.Fatalf("status = %q", disabled.Status)
	}
	if _, err := notify.UpdateOnCallStatus(ctx, ns, created.UID, "BROKEN"); !merr.IsInvalid(err) {
		t.Fatalf("invalid status error = %v", err)
	}

	// 同名资源在未删除记录上唯一。
	if _, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{NamespaceUID: ns, Name: "primary-renamed"}); !merr.IsConflict(err) {
		t.Fatalf("duplicate name error = %v, want conflict", err)
	}
	// 非法时区和非法时段在保存时拒绝。
	if _, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{NamespaceUID: ns, Name: "bad-zone", Timezone: "Mars/Olympus"}); !merr.IsInvalid(err) {
		t.Fatalf("invalid timezone error = %v", err)
	}
	if _, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: ns, Name: "bad-window",
		Groups: []bo.NotificationOnCallGroup{{GroupUID: groupB.UID, Windows: []bo.NotificationOnCallWindow{
			{Weekdays: []string{"MONDAY"}, Start: "09:00", End: "09:00"},
		}}},
	}); !merr.IsInvalid(err) {
		t.Fatalf("invalid window error = %v", err)
	}

	if err := notify.DeleteOnCall(ctx, ns, created.UID); err != nil {
		t.Fatal(err)
	}
	if _, err := notify.GetOnCall(ctx, ns, created.UID); !merr.IsNotFound(err) {
		t.Fatalf("get deleted on-call error = %v, want not found", err)
	}
	if err := notify.DeleteOnCall(ctx, ns, created.UID); !merr.IsNotFound(err) {
		t.Fatalf("delete twice error = %v, want not found", err)
	}
}

func TestOnCallDispatchWindowsAndDedup(t *testing.T) {
	ctx, store, notify, events, ns := onCallTestEnvironment(t)
	groupActive := newOnCallTestGroup(t, ctx, notify, ns, "active")
	groupInactive := newOnCallTestGroup(t, ctx, notify, ns, "inactive")

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	localNow := time.Now().In(location)
	activeOnCall, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: ns, Name: "active-oncall", Labels: `{"env":"prod"}`,
		LevelUIDs: []string{"lv1"}, DatasourceUIDs: []string{"ds1"}, Timezone: "Asia/Shanghai",
		Groups: []bo.NotificationOnCallGroup{{GroupUID: groupActive.UID, Windows: onCallWindowAround(localNow, -1)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	inactiveOnCall, err := notify.CreateOnCall(ctx, bo.NotificationOnCall{
		NamespaceUID: ns, Name: "inactive-oncall", Labels: `{"env":"prod"}`,
		LevelUIDs: []string{"lv1"}, DatasourceUIDs: []string{"ds1"}, Timezone: "Asia/Shanghai",
		Groups: []bo.NotificationOnCallGroup{{GroupUID: groupInactive.UID, Windows: onCallWindowAround(localNow, 2)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(inactiveOnCall.Groups) != 1 || len(inactiveOnCall.Groups[0].Windows) != 1 {
		t.Fatalf("inactive on-call windows = %+v", inactiveOnCall.Groups)
	}

	// 每个场景用独立规则，避免上一轮命中的指纹被判定恢复而额外分发 RECOVERED 事件。
	slot := time.Now().UTC().Truncate(time.Minute)
	fire := func(ruleName string, slotOffset time.Duration, datasourceUID, levelUID, labels string) {
		t.Helper()
		rule := bo.AlertRule{UID: "rule-" + ruleName + "-" + ns, NamespaceUID: ns, Expr: "up", Summary: "cpu"}
		if err := events.ApplyReport(ctx, rule, bo.EvalReport{
			RuleUID: rule.UID, DatasourceUID: datasourceUID, EvalSlot: slot.Add(slotOffset),
			Hits: []bo.EvalHit{{
				Fingerprint: ruleName, Labels: labels, Value: "1", LevelUID: levelUID, Summary: "fire",
			}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	groupTasks := func(groupUID string) []*ent.NotificationTask {
		t.Helper()
		rows, err := store.client.NotificationTask.Query().Where(
			notificationtask.NamespaceUIDEQ(ns),
			notificationtask.GroupUIDEQ(groupUID),
		).Order(ent.Asc(notificationtask.FieldID)).All(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}

	// 命中 ACTIVE OnCall 且当前时刻在时段内：生成一条任务，subscription_uid 留空。
	fire("main", 0, "ds1", "lv1", `{"env":"prod"}`)
	activeTasks := groupTasks(groupActive.UID)
	if len(activeTasks) != 1 {
		t.Fatalf("active window tasks = %d, want 1", len(activeTasks))
	}
	task := activeTasks[0]
	if task.SubscriptionUID != "" {
		t.Fatalf("on-call task subscription_uid = %q, want empty", task.SubscriptionUID)
	}
	if task.Phase != notificationtask.PhaseFIRING || task.Status != notificationtask.StatusPENDING || task.RetriesLeft != 3 {
		t.Fatalf("task state = %+v", task)
	}
	if task.ChannelUID == "" || task.TemplateUID == "" || task.Subject != "alert" || task.Body != "{{.summary}}" || task.Content != "{{.summary}}" {
		t.Fatalf("task target fields = %+v", task)
	}
	if task.Params == "" {
		t.Fatal("task params is empty")
	}
	// 当前时刻在时段外：不生成任务。
	if tasks := groupTasks(groupInactive.UID); len(tasks) != 0 {
		t.Fatalf("inactive window created tasks: %d", len(tasks))
	}

	// 标签、告警级别、数据源任一不满足都不发送。
	fire("label-mismatch", time.Minute, "ds1", "lv1", `{"env":"dev"}`)
	fire("datasource-mismatch", 2*time.Minute, "ds2", "lv1", `{"env":"prod"}`)
	fire("level-mismatch", 3*time.Minute, "ds1", "lv2", `{"env":"prod"}`)
	if tasks := groupTasks(groupActive.UID); len(tasks) != 1 {
		t.Fatalf("condition mismatch created tasks: %d", len(tasks))
	}

	// 通知组停用后同样不再产生任务。
	if _, err := notify.UpdateGroupStatus(ctx, ns, groupActive.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	fire("disabled-group", 4*time.Minute, "ds1", "lv1", `{"env":"prod"}`)
	if tasks := groupTasks(groupActive.UID); len(tasks) != 1 {
		t.Fatalf("disabled group created tasks: %d", len(tasks))
	}
	if _, err := notify.UpdateGroupStatus(ctx, ns, groupActive.UID, "ACTIVE"); err != nil {
		t.Fatal(err)
	}

	// 停用 OnCall 后不再产生任务。
	if _, err := notify.UpdateOnCallStatus(ctx, ns, activeOnCall.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	fire("disabled-oncall", 5*time.Minute, "ds1", "lv1", `{"env":"prod"}`)
	if tasks := groupTasks(groupActive.UID); len(tasks) != 1 {
		t.Fatalf("disabled on-call created tasks: %d", len(tasks))
	}
	if _, err := notify.UpdateOnCallStatus(ctx, ns, activeOnCall.UID, "ACTIVE"); err != nil {
		t.Fatal(err)
	}

	// 同一通知组同时被订阅和 OnCall 命中时，唯一约束合并且订阅标识优先。
	subscription, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "prod-sub", Labels: "{}", GroupUIDs: []string{groupActive.UID},
	})
	if err != nil {
		t.Fatal(err)
	}
	fire("dedup", 6*time.Minute, "ds1", "lv1", `{"env":"prod"}`)
	activeTasks = groupTasks(groupActive.UID)
	if len(activeTasks) != 2 {
		t.Fatalf("duplicate suppression failed, tasks = %d, want 2", len(activeTasks))
	}
	newest := activeTasks[len(activeTasks)-1]
	if newest.SubscriptionUID != subscription.UID {
		t.Fatalf("deduplicated task subscription_uid = %q, want %q", newest.SubscriptionUID, subscription.UID)
	}
}
