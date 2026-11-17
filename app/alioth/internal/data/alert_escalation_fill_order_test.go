package data

import (
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventescalation"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

// TestFillPendingEscalationsTakesOldestFirst 确认补行按事件开始时间从旧到新取。
// 没有排序时取到的是表的物理顺序（即写入顺序），窗口里超过上限的事件里最新那批
// 会一直占住名额、更旧的事件反而补不到行——这里按「新、中、旧」的写入顺序造数据，
// 上限设为 2 后只有最旧的两条应该拿到行。
func TestFillPendingEscalationsTakesOldestFirst(t *testing.T) {
	store, ctx := newEscalationTestStore(t)
	namespace := "ns-escalation-order-" + time.Now().Format("150405.000000")
	t.Cleanup(cleanupEscalationNamespace(store, namespace))
	repo := NewAlertEscalationRepo(store, slog.Default())
	groupUID, _ := escalationTestFixture(ctx, t, store, namespace)
	now := time.Now().UTC()

	policy, err := repo.CreatePolicy(ctx, bo.AlertEscalationPolicy{
		NamespaceUID: namespace,
		Name:         "顺序校验",
		Labels:       `{}`,
		Steps: []bo.AlertEscalationStep{
			{DelayMinutes: 5, NotificationGroupUIDs: []string{groupUID}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 调度只升级策略创建之后开始的事件；把创建时间前移，让它覆盖这三条事件。
	if _, err := store.db.ExecContext(ctx,
		`UPDATE alert_escalation_policies SET created_at = $1 WHERE uid = $2`,
		now.Add(-24*time.Hour), policy.UID); err != nil {
		t.Fatal(err)
	}

	oldest := id.New()
	middle := id.New()
	newest := id.New()
	// 故意按「最新、中间、最旧」写入：物理顺序与 starts_at 顺序相反。
	escalationTestEvent(ctx, t, store, namespace, newest, "fp-order-newest", now.Add(-16*time.Minute))
	escalationTestEvent(ctx, t, store, namespace, middle, "fp-order-middle", now.Add(-18*time.Minute))
	escalationTestEvent(ctx, t, store, namespace, oldest, "fp-order-oldest", now.Add(-20*time.Minute))

	// 直接用小上限调实现，验证取数顺序；公开方法固定 500 条。
	impl, ok := repo.(*alertEscalationRepo)
	if !ok {
		t.Fatalf("unexpected repo type %T", repo)
	}
	if _, err := impl.fillPendingEscalations(ctx, now.Add(-24*time.Hour), now, 2); err != nil {
		t.Fatalf("fill pending: %v", err)
	}

	hasRows := func(eventUID string) bool {
		t.Helper()
		found, err := store.client.AlertEventEscalation.Query().Where(
			alerteventescalation.EventUIDEQ(eventUID),
			alerteventescalation.PolicyUIDEQ(policy.UID),
		).Exist(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return found
	}
	if !hasRows(oldest) || !hasRows(middle) {
		t.Fatal("最旧的两条事件应该先拿到升级行")
	}
	if hasRows(newest) {
		t.Fatal("超过上限的最新事件不应在这一轮拿行")
	}

	// 下一轮把上限放开，剩余事件补上；已经补过行的不会重复。
	if _, err := repo.FillPendingEscalations(ctx, now.Add(-24*time.Hour), now); err != nil {
		t.Fatalf("fill pending second pass: %v", err)
	}
	if !hasRows(newest) {
		t.Fatal("放宽上限后最新事件应补到行")
	}
	count, err := store.client.AlertEventEscalation.Query().Where(
		alerteventescalation.PolicyUIDEQ(policy.UID),
	).Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("escalation rows = %d, want 3", count)
	}
}
