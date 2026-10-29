package data

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/proberesult"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestProbeEventFiringAndRecovery 校验探测事件的字段口径：没有数据源、expr 是目标
// 地址、labels 带目标组名与目标标识；同一目标空回写会恢复事件，删除目标时按
// probe_target_uid 收口。需要 ALIOTH_TEST_DSN 指向非 dev 库；没有配置时跳过。
func TestProbeEventFiringAndRecovery(t *testing.T) {
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

	namespace := "ns-probe-" + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(namespace)).Exec(ctx)
		_ = store.history.dropTenant(context.Background(), namespace)
	})

	events := NewAlertEventRepo(store, slog.Default())
	group := bo.ProbeTargetGroup{
		UID: "group-" + namespace, Name: "核心接口组", Driver: bo.ProbeDriverHTTP,
		Labels: `{"env":"prod"}`,
	}
	target := bo.ProbeTarget{
		UID: "target-" + namespace, GroupUID: group.UID, Spec: "https://api.example.com/healthz",
	}
	rule := bo.AlertRule{
		UID: "rule-" + namespace, NamespaceUID: namespace, Name: "API 健康检查",
		Driver: bo.ProbeDriverHTTP, ProbeMetric: bo.ProbeMetricUp,
	}
	labels := map[string]string{
		"probe_target_uid": target.UID,
		"probe_target":     target.Spec,
		"probe_kind":       "HTTP",
		"env":              "prod",
	}
	encoded, err := json.Marshal(labels)
	if err != nil {
		t.Fatal(err)
	}
	slot := time.Now().UTC().Truncate(time.Second)
	hit := bo.EvalHit{
		Fingerprint: "fp-" + namespace, Labels: string(encoded), Value: "0",
		LevelUID: "lvl-critical", Summary: target.Spec + " 探测失败", Description: "connection refused",
	}
	if _, err := events.ApplyProbeHits(ctx, rule, group, target, slot, []bo.EvalHit{hit}); err != nil {
		t.Fatal(err)
	}

	row, err := store.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespace), alertevent.FingerprintEQ(hit.Fingerprint),
	).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.DatasourceUID != "" {
		t.Fatalf("probe event should not carry datasource uid, got %q", row.DatasourceUID)
	}
	if row.Expr != target.Spec {
		t.Fatalf("expr should be the target spec, got %q", row.Expr)
	}
	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(row.Labels), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["datasource"] != group.Name {
		t.Fatalf("labels.datasource = %q, want %q", parsed["datasource"], group.Name)
	}
	if parsed["probe_target_uid"] != target.UID {
		t.Fatalf("labels.probe_target_uid = %q", parsed["probe_target_uid"])
	}

	// 列表的「数据源」列在 datasource_uid 为空时用标签兜底。
	listed, _, err := events.ListEvents(ctx, namespace, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 realtime event, got %d", len(listed))
	}
	if listed[0].DatasourceName != group.Name {
		t.Fatalf("listed datasource name = %q, want %q", listed[0].DatasourceName, group.Name)
	}

	// 同一目标空回写（探测恢复）应把事件置为恢复。
	if _, err := events.ApplyProbeHits(ctx, rule, group, target, slot.Add(30*time.Second), nil); err != nil {
		t.Fatal(err)
	}
	// 恢复后事件归档到历史分片，主库行与 firing 占位都应消失。
	if _, err := store.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespace), alertevent.FingerprintEQ(hit.Fingerprint),
	).Only(ctx); !ent.IsNotFound(err) {
		t.Fatalf("event should be archived after recovery, err = %v", err)
	}
	if _, err := store.client.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(namespace), alerteventfiring.FingerprintEQ(hit.Fingerprint),
	).Only(ctx); !ent.IsNotFound(err) {
		t.Fatalf("firing placeholder should be removed, err = %v", err)
	}
}

// TestSaveProbeResultWithoutRule 立即探测没有规则，写入的 probe_results 允许
// rule_uid 为空；规则判定的记录才带规则标识。
func TestSaveProbeResultWithoutRule(t *testing.T) {
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
	namespace := "ns-probe-now-" + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		_, _ = store.client.ProbeResult.Delete().Where(proberesult.NamespaceUIDEQ(namespace)).Exec(ctx)
	})

	repo := NewProbeRepo(store, slog.Default())
	slot := time.Now().UTC().Truncate(time.Second)
	if _, err := repo.SaveProbeResults(ctx, []bo.ProbeResult{{
		NamespaceUID: namespace, GroupUID: "group-now", TargetUID: "target-now",
		TargetSpec: "http://127.0.0.1:8002/health", Slot: slot, ProbedAt: slot,
		Value: "12", DurationMS: 12, StatusCode: 200, Executor: "manual",
	}}); err != nil {
		t.Fatalf("probe-now result without rule should be saved: %v", err)
	}
	row, err := store.client.ProbeResult.Query().Where(
		proberesult.NamespaceUIDEQ(namespace), proberesult.TargetUIDEQ("target-now"),
	).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.RuleUID != "" {
		t.Fatalf("probe-now row rule_uid = %q, want empty", row.RuleUID)
	}
}

// TestSaveProbeResultsReportsOnlyNewTargets 同一槽位重复回写不会重复判定和通知：
// SaveProbeResults 只把真正新增的目标返给调用方。
func TestSaveProbeResultsReportsOnlyNewTargets(t *testing.T) {
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
	namespace := "ns-probe-dedup-" + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		_, _ = store.client.ProbeResult.Delete().Where(proberesult.NamespaceUIDEQ(namespace)).Exec(ctx)
	})
	repo := NewProbeRepo(store, slog.Default())
	slot := time.Now().UTC().Truncate(time.Second)
	result := bo.ProbeResult{
		NamespaceUID: namespace, GroupUID: "group-dedup", TargetUID: "target-dedup",
		TargetSpec: "http://127.0.0.1:8002/health", Slot: slot, ProbedAt: slot,
		OK: true, Value: "1", DurationMS: 9, StatusCode: 200,
	}
	first, err := repo.SaveProbeResults(ctx, []bo.ProbeResult{result})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := first["target-dedup"]; !ok {
		t.Fatalf("first write should report the target as new: %v", first)
	}
	second, err := repo.SaveProbeResults(ctx, []bo.ProbeResult{result})
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("duplicate write must report no new targets, got %v", second)
	}
	probed, err := repo.ProbingTargetUIDs(ctx, namespace, slot, []string{"target-dedup"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := probed["target-dedup"]; !ok {
		t.Fatal("slot should be marked as probed")
	}
}
