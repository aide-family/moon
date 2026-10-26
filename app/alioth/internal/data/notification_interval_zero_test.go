package data

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// TestFiringSuppressedHonoursZeroInterval 间隔 0 表示不限制：
// 同一指纹已有 PENDING FIRING 任务时也不抑制；正间隔才抑制。
// 直接在仓库层断言，避免依赖“重复 FIRING 一定产生新事件”的上游状态机假设。
func TestFiringSuppressedHonoursZeroInterval(t *testing.T) {
	ctx, store, notify, events, ns := onCallTestEnvironment(t)
	t.Cleanup(func() { _ = store.history.dropTenant(context.Background(), ns) })
	datasourceUID := "ds-" + ns
	registerReportCleanup(t, ctx, store, datasourceUID)
	group := newNotificationTestEmailGroup(t, ctx, notify, ns, "zero", "zero@example.com")
	if _, err := notify.CreateSubscription(ctx, bo.NotificationSubscription{
		NamespaceUID: ns, Name: "zero-sub", Labels: "{}", RepeatIntervalSeconds: 0,
		GroupUIDs: []string{group.UID},
	}); err != nil {
		t.Fatal(err)
	}
	slot := time.Now().UTC().Truncate(time.Minute)
	applyFiringReport(t, ctx, events, ns, datasourceUID, "fp-zero", slot)
	if tasks := tasksByFingerprint(t, ctx, store, ns, "fp-zero"); len(tasks) == 0 {
		t.Fatal("expected the first firing to create a task")
	}

	tx, err := store.client.Tx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	event := notifyEvent{NamespaceUID: ns, Fingerprint: "fp-zero"}

	suppressed, err := firingSuppressed(ctx, tx, event, 0)
	if err != nil {
		t.Fatal(err)
	}
	if suppressed {
		t.Fatal("zero interval must not suppress a firing")
	}

	suppressed, err = firingSuppressed(ctx, tx, event, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !suppressed {
		t.Fatal("a positive interval must suppress a recent firing")
	}
}
