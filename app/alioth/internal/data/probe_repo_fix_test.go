package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/proberesult"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetarget"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// newProbeFixStore 连接集成测试库并应用 schema，模式与 probe_event_integration_test.go 一致。
func newProbeFixStore(t *testing.T) (*Dependencies, context.Context) {
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

// cleanupProbeFixNamespace 清掉本用例写入的探测数据。
func cleanupProbeFixNamespace(store *Dependencies, namespace string) func() {
	return func() {
		ctx := context.Background()
		_, _ = store.client.ProbeResult.Delete().Where(proberesult.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.ProbeTarget.Delete().Where(probetarget.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.ProbeTargetGroup.Delete().Where(probetargetgroup.NamespaceUIDEQ(namespace)).Exec(ctx)
	}
}

// TestUpdateProbeTargetGroupKeepsNewTargets 整体替换目标列表时，payload 里 uid 为空的
// 新目标先创建、再被「不在 keep 的现有行」清扫逻辑误删，导致页面加完地址后数量不变。
func TestUpdateProbeTargetGroupKeepsNewTargets(t *testing.T) {
	store, ctx := newProbeFixStore(t)
	namespace := "ns-probe-fix-" + time.Now().Format("150405.000000")
	t.Cleanup(cleanupProbeFixNamespace(store, namespace))
	repo := NewProbeRepo(store, slog.Default())

	created, err := repo.CreateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		NamespaceUID: namespace, Name: "核心接口组",
		Labels: "{}", ExecutorLabels: "{}", Request: "{}",
		Targets: []bo.ProbeTarget{{
			Spec: "https://old.example.com/health", Status: "ACTIVE",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.TargetCount != 1 || len(created.Targets) != 1 {
		t.Fatalf("created target count = %d/%d, want 1", created.TargetCount, len(created.Targets))
	}
	oldUID := created.Targets[0].UID

	// payload 保留旧目标并追加一个 uid 为空的新目标。
	updated, err := repo.UpdateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		UID: created.UID, NamespaceUID: namespace, Name: "核心接口组",
		Labels: "{}", ExecutorLabels: "{}", Request: "{}",
		Targets: []bo.ProbeTarget{
			{UID: oldUID, Spec: "https://old.example.com/healthz", Status: "ACTIVE"},
			{Spec: "https://new.example.com/health", Status: "ACTIVE", Sort: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.TargetCount != 2 || len(updated.Targets) != 2 {
		t.Fatalf("after adding a target, target count = %d/%d, want 2", updated.TargetCount, len(updated.Targets))
	}
	specs := make(map[string]string, len(updated.Targets))
	for _, target := range updated.Targets {
		specs[target.Spec] = target.UID
		if target.UID == "" {
			t.Fatalf("target %q should get a generated uid", target.Spec)
		}
	}
	if specs["https://old.example.com/healthz"] != oldUID {
		t.Fatalf("old target should be updated in place, got %q", specs["https://old.example.com/healthz"])
	}
	newUID := specs["https://new.example.com/health"]
	if newUID == "" {
		t.Fatalf("new target should be kept, specs = %v", specs)
	}
	// 新目标行必须没有 deleted_at；被误删时 Get 系列会把它过滤掉，直接查库确认。
	newRow, err := store.client.ProbeTarget.Query().Where(probetarget.UIDEQ(newUID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if newRow.DeletedAt != nil {
		t.Fatalf("new target %q should not be soft-deleted, deleted_at = %v", newUID, newRow.DeletedAt)
	}

	// 设计语义保留：payload 里不带的旧目标仍要软删。
	pruned, err := repo.UpdateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		UID: created.UID, NamespaceUID: namespace, Name: "核心接口组",
		Labels: "{}", ExecutorLabels: "{}", Request: "{}",
		Targets: []bo.ProbeTarget{
			{UID: oldUID, Spec: "https://old.example.com/healthz", Status: "ACTIVE"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pruned.TargetCount != 1 || len(pruned.Targets) != 1 {
		t.Fatalf("after dropping a target, target count = %d/%d, want 1", pruned.TargetCount, len(pruned.Targets))
	}
	if pruned.Targets[0].UID != oldUID {
		t.Fatalf("remaining target uid = %q, want %q", pruned.Targets[0].UID, oldUID)
	}
	droppedRow, err := store.client.ProbeTarget.Query().Where(probetarget.UIDEQ(newUID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if droppedRow.DeletedAt == nil {
		t.Fatalf("target %q missing from payload should be soft-deleted", newUID)
	}
}

// TestProbeResultsFillTargetSpec 探测记录列表与最新结果都要能显示目标地址；
// 历史目标被软删后仍要能补出 spec，不能让记录里的地址列变成空。
func TestProbeResultsFillTargetSpec(t *testing.T) {
	store, ctx := newProbeFixStore(t)
	namespace := "ns-probe-spec-" + time.Now().Format("150405.000000")
	t.Cleanup(cleanupProbeFixNamespace(store, namespace))
	repo := NewProbeRepo(store, slog.Default())

	const spec = "https://api.example.com/healthz"
	group, err := repo.CreateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		NamespaceUID: namespace, Name: "spec 组",
		Labels: "{}", ExecutorLabels: "{}", Request: "{}",
		Targets: []bo.ProbeTarget{{Spec: spec, Status: "ACTIVE"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(group.Targets) != 1 {
		t.Fatalf("created targets = %d, want 1", len(group.Targets))
	}
	targetUID := group.Targets[0].UID

	slot := time.Now().UTC().Truncate(time.Second)
	if _, err := repo.SaveProbeResults(ctx, []bo.ProbeResult{{
		NamespaceUID: namespace, GroupUID: group.UID, TargetUID: targetUID,
		Slot: slot, ProbedAt: slot, OK: true, Value: "1", StatusCode: 200,
		DurationMS: 12, Executor: "manual",
	}}); err != nil {
		t.Fatal(err)
	}

	listed, total, err := repo.ListProbeTargetResults(ctx, namespace, group.UID, "", 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(listed) != 1 {
		t.Fatalf("listed results = %d/%d, want 1", total, len(listed))
	}
	if listed[0].TargetSpec != spec {
		t.Fatalf("ListProbeTargetResults target_spec = %q, want %q", listed[0].TargetSpec, spec)
	}

	latest, err := repo.LatestProbeResults(ctx, []string{targetUID})
	if err != nil {
		t.Fatal(err)
	}
	latestItem, ok := latest[targetUID]
	if !ok {
		t.Fatalf("LatestProbeResults missing target %q", targetUID)
	}
	if latestItem.TargetSpec != spec {
		t.Fatalf("LatestProbeResults target_spec = %q, want %q", latestItem.TargetSpec, spec)
	}

	// 目标从组里移除后行会被软删，探测记录仍在，地址列不能因此变空。
	if _, err := repo.UpdateProbeTargetGroup(ctx, bo.ProbeTargetGroup{
		UID: group.UID, NamespaceUID: namespace, Name: "spec 组",
		Labels: "{}", ExecutorLabels: "{}", Request: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	latestDeleted, err := repo.LatestProbeResults(ctx, []string{targetUID})
	if err != nil {
		t.Fatal(err)
	}
	deletedItem, ok := latestDeleted[targetUID]
	if !ok {
		t.Fatalf("LatestProbeResults missing soft-deleted target %q", targetUID)
	}
	if deletedItem.TargetSpec != spec {
		t.Fatalf("LatestProbeResults target_spec after soft delete = %q, want %q", deletedItem.TargetSpec, spec)
	}
	listedDeleted, totalDeleted, err := repo.ListProbeTargetResults(ctx, namespace, group.UID, "", 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if totalDeleted != 1 || len(listedDeleted) != 1 {
		t.Fatalf("listed results after soft delete = %d/%d, want 1", totalDeleted, len(listedDeleted))
	}
	if listedDeleted[0].TargetSpec != spec {
		t.Fatalf("ListProbeTargetResults target_spec after soft delete = %q, want %q", listedDeleted[0].TargetSpec, spec)
	}
}
