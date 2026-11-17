package data

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestAvailabilityRaceKeepsFailedBatch 覆盖多个副本同时给同一数据源记失败批次的场景：
// 抢同一条可用性占位失败的一方只能回滚自己的事务，已经记账的失败批次必须留在库里。
// 放在同一个事务里时，冲突方会把失败批次一起回滚，阈值计数少一次，可用性告警可能
// 永远凑不够阈值。
func TestAvailabilityRaceKeepsFailedBatch(t *testing.T) {
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
	namespace := "ns-availability-race-" + time.Now().Format("150405.000000")
	sources := NewDatasourceRepo(store, slog.Default())
	created, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: namespace, Name: "prom-race", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: "http://127.0.0.1:9090", Labels: `{}`,
		AvailabilityAlertEnabled: true, FailureWindowMinutes: 5, FailureThreshold: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = store.client.Datasource.Delete().Where(datasource.UIDEQ(created.UID)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.DatasourceUIDEQ(created.UID)).Exec(ctx)
		_, _ = store.client.AlertEventFiring.Delete().Where(alerteventfiring.NamespaceUIDEQ(namespace)).Exec(ctx)
		_, _ = store.client.NotificationTask.Delete().Where(notificationtask.NamespaceUIDEQ(namespace)).Exec(ctx)
		_ = store.history.dropTenant(context.Background(), namespace)
	})

	events := NewAlertEventRepo(store, slog.Default())
	base := time.Now().UTC().Truncate(time.Second)
	const rounds = 5
	for round := 0; round < rounds; round++ {
		// 每轮换一个评测槽并换两条规则：唯一约束允许同一数据源不同规则、同槽各记一次失败。
		ruleA := bo.AlertRule{UID: "rule-race-a-" + namespace, NamespaceUID: namespace, DatasourceUID: created.UID}
		ruleB := bo.AlertRule{UID: "rule-race-b-" + namespace, NamespaceUID: namespace, DatasourceUID: created.UID}
		slot := base.Add(time.Duration(round) * time.Second)
		var wg sync.WaitGroup
		errs := make([]error, 2)
		start := make(chan struct{})
		for index, rule := range []bo.AlertRule{ruleA, ruleB} {
			wg.Add(1)
			go func(index int, rule bo.AlertRule) {
				defer wg.Done()
				<-start
				errs[index] = events.ApplyReport(ctx, rule, bo.EvalReport{
					RuleUID: rule.UID, DatasourceUID: created.UID, EvalSlot: slot,
					Failed: true, FailureMessage: "upstream query failed",
				})
			}(index, rule)
		}
		close(start)
		wg.Wait()
		for index, err := range errs {
			if err != nil {
				t.Fatalf("round %d report %d: %v", round, index, err)
			}
		}
	}

	failed := store.client.AlertEvalBatch.Query().Where(
		alertevalbatch.DatasourceUIDEQ(created.UID),
		alertevalbatch.FailedEQ(true),
	).CountX(ctx)
	if failed != rounds*2 {
		t.Fatalf("failed batches = %d, want %d", failed, rounds*2)
	}
	// 同一数据源只有一条可用性事件；抢占失败的副本不应留下第二条。
	firing := store.client.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(namespace),
	).CountX(ctx)
	if firing != 1 {
		t.Fatalf("firing placeholders = %d, want 1", firing)
	}
}
