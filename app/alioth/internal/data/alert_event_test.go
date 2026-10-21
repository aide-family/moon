package data

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestAlertEventReport(t *testing.T) {
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
	events := NewAlertEventRepo(store, slog.Default())
	ns := "ns-event-" + time.Now().Format("150405.000")
	rule := bo.AlertRule{UID: "rule-" + ns, NamespaceUID: ns, DatasourceUID: "ds", Expr: "up", Summary: "cpu"}
	slot := time.Now().UTC().Truncate(time.Second)
	report := bo.EvalReport{
		RuleUID: rule.UID, EvalSlot: slot,
		Hits: []bo.EvalHit{{Fingerprint: "fp-1", Labels: `{"job":"api-1"}`, Value: "1", Summary: "cpu"}},
	}
	t.Cleanup(func() {
		_, _ = store.client.AlertEventFiring.Delete().Where(alerteventfiring.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.RuleUIDEQ(rule.UID)).Exec(ctx)
	})

	if err := events.ApplyReport(ctx, rule, report); err != nil {
		t.Fatal(err)
	}
	if err := events.ApplyReport(ctx, rule, report); err != nil {
		t.Fatal(err)
	}
	firing, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(firing) != 1 || firing[0].Status != "FIRING" {
		t.Fatalf("firing = %+v", firing)
	}

	// 告警页面的标签值按完整匹配正则过滤；精确值、通配和计数使用同一套语义。
	regexFilter := bo.AlertEventListFilter{View: "REALTIME", LabelsJSON: `{"job":"api-.+"}`}
	regexEvents, regexTotal, err := events.ListEvents(ctx, ns, regexFilter, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(regexEvents) != 1 || regexTotal != 1 {
		t.Fatalf("regex events = %+v, total = %d", regexEvents, regexTotal)
	}
	if count, err := events.CountEvents(ctx, ns, regexFilter); err != nil || count != 1 {
		t.Fatalf("regex count = %d, err = %v", count, err)
	}
	if unmatched, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{
		View: "REALTIME", LabelsJSON: `{"job":"web-.*"}`,
	}, 1, 20); err != nil || len(unmatched) != 0 {
		t.Fatalf("non-matching regex events = %+v, err = %v", unmatched, err)
	}
	if exact, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{
		View: "REALTIME", LabelsJSON: `{"job":"api-1"}`,
	}, 1, 20); err != nil || len(exact) != 1 {
		t.Fatalf("exact match events = %+v, err = %v", exact, err)
	}
	// 完整匹配：api 不能命中 api-1。
	if partial, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{
		View: "REALTIME", LabelsJSON: `{"job":"api"}`,
	}, 1, 20); err != nil || len(partial) != 0 {
		t.Fatalf("partial match events = %+v, err = %v", partial, err)
	}
	// 这些写法 PostgreSQL 的 ARE 不支持或语义不同，必须由 Go RE2 统一处理。
	for _, pattern := range []string{`\Qapi-1\E`, `\bapi-\d+`, `api-\w+`} {
		labelsJSON, err := json.Marshal(map[string]string{"job": pattern})
		if err != nil {
			t.Fatal(err)
		}
		matched, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{
			View: "REALTIME", LabelsJSON: string(labelsJSON),
		}, 1, 20)
		if err != nil || len(matched) != 1 {
			t.Fatalf("regex %q events = %+v, err = %v", pattern, matched, err)
		}
	}

	recovered := bo.EvalReport{RuleUID: rule.UID, EvalSlot: slot.Add(time.Second)}
	if err := events.ApplyReport(ctx, rule, recovered); err != nil {
		t.Fatal(err)
	}
	if err := events.ApplyReport(ctx, rule, recovered); err != nil {
		t.Fatal(err)
	}
	history, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "HISTORY"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Status != "RECOVERED" {
		t.Fatalf("history = %+v", history)
	}
	again, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Fatalf("realtime after recover = %+v", again)
	}
}

// TestAlertEventDatasourceLabels 校验事件标签附带数据源名称和数据源自身标签，
// 并且详情读取时能补出规则名称、规则组和级别名称。
func TestAlertEventDatasourceLabels(t *testing.T) {
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
	events := NewAlertEventRepo(store, slog.Default())
	ns := "ns-labels-" + time.Now().Format("150405.000")
	datasourceUID := "ds-" + ns
	ruleUID := "rule-" + ns
	levelUID := "level-" + ns
	datasourceLevelUID := "datasource-level-" + ns
	groupUID := "group-" + ns
	slot := time.Now().UTC().Truncate(time.Second)
	rule := bo.AlertRule{UID: ruleUID, NamespaceUID: ns, DatasourceUID: datasourceUID, Expr: "up", Summary: "cpu"}
	t.Cleanup(func() {
		_, _ = store.client.AlertEventFiring.Delete().Where(alerteventfiring.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.RuleUIDEQ(ruleUID)).Exec(ctx)
		_, _ = store.client.AlertRule.Delete().Where(alertrule.UIDEQ(ruleUID)).Exec(ctx)
		_, _ = store.client.AlertDictionary.Delete().Where(alertdictionary.UIDIn(levelUID, datasourceLevelUID)).Exec(ctx)
		_, _ = store.client.Datasource.Delete().Where(datasource.UIDEQ(datasourceUID)).Exec(ctx)
	})

	if err := store.client.Datasource.Create().
		SetUID(datasourceUID).
		SetNamespaceUID(ns).
		SetName("prom-prod").
		SetType("METRICS").
		SetDriver("METRICS_PROMETHEUS").
		SetEndpoint("http://127.0.0.1:9090").
		SetLevelUID(datasourceLevelUID).
		SetLabels(`{"env":"prod","region":"cn"}`).
		Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.client.AlertRule.Create().
		SetUID(ruleUID).
		SetNamespaceUID(ns).
		SetGroupUID(groupUID).
		SetName("CPU high").
		SetType("METRICS").
		SetDriver("METRICS_PROMETHEUS").
		SetDatasourceUID(datasourceUID).
		Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.client.AlertDictionary.Create().
		SetUID(levelUID).
		SetNamespaceUID(ns).
		SetName("P1").
		SetType("ALERT").
		Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if err := store.client.AlertDictionary.Create().
		SetUID(datasourceLevelUID).
		SetNamespaceUID(ns).
		SetName("P2").
		SetType("ALERT").
		Exec(ctx); err != nil {
		t.Fatal(err)
	}

	// 指标标签里的 env 优先，不被数据源标签覆盖；数据源标签只补缺。
	report := bo.EvalReport{
		RuleUID: ruleUID, DatasourceUID: datasourceUID, EvalSlot: slot,
		Hits: []bo.EvalHit{{
			Fingerprint: "fp-1", Labels: `{"job":"api","env":"staging"}`,
			Value: "1", LevelUID: levelUID,
		}},
	}
	if err := events.ApplyReport(ctx, rule, report); err != nil {
		t.Fatal(err)
	}
	firing, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(firing) != 1 {
		t.Fatalf("firing = %+v", firing)
	}
	var labels map[string]string
	if err := json.Unmarshal([]byte(firing[0].Labels), &labels); err != nil {
		t.Fatalf("labels = %q, err = %v", firing[0].Labels, err)
	}
	for key, want := range map[string]string{
		"job": "api", "env": "staging", "region": "cn", "datasource": "prom-prod",
		"alert_level": "P1", "datasource_level": "P2",
	} {
		if labels[key] != want {
			t.Fatalf("labels[%q] = %q, want %q (labels = %s)", key, labels[key], want, firing[0].Labels)
		}
	}
	if firing[0].DatasourceName != "prom-prod" || firing[0].RuleName != "CPU high" ||
		firing[0].RuleGroupUID != groupUID || firing[0].LevelName != "P1" {
		t.Fatalf("event names = %+v", firing[0])
	}
}

// TestAlertEventSuppression 校验按秒抑制、抑制到期回到实时列表，
// 以及再次命中时清掉已过期的抑制状态。
func TestAlertEventSuppression(t *testing.T) {
	dsn := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewData(
		&conf.Data{Database: &conf.Data_Database{Source: dsn}},
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
	events := NewAlertEventRepo(store, slog.Default())
	ns := "ns-suppress-" + time.Now().Format("150405.000")
	rule := bo.AlertRule{UID: "rule-" + ns, NamespaceUID: ns, DatasourceUID: "ds", Expr: "up", Summary: "cpu"}
	report := bo.EvalReport{
		RuleUID: rule.UID, EvalSlot: time.Now().UTC().Truncate(time.Second),
		Hits: []bo.EvalHit{{Fingerprint: "fp-1", Labels: `{"job":"api-1"}`, Value: "1", Summary: "cpu"}},
	}
	t.Cleanup(func() {
		_, _ = store.client.AlertEventFiring.Delete().Where(alerteventfiring.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvalBatch.Delete().Where(alertevalbatch.RuleUIDEQ(rule.UID)).Exec(ctx)
	})
	if err := events.ApplyReport(ctx, rule, report); err != nil {
		t.Fatal(err)
	}
	firing, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil || len(firing) != 1 {
		t.Fatalf("firing = %+v, err = %v", firing, err)
	}

	// 抑制 1 小时：离开实时列表，进入历史，统计计入抑制。
	until := time.Now().UTC().Add(time.Hour)
	if err := events.Suppress(ctx, ns, firing[0].UID, &until); err != nil {
		t.Fatal(err)
	}
	if realtime, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20); err != nil || len(realtime) != 0 {
		t.Fatalf("realtime after suppress = %+v, err = %v", realtime, err)
	}
	history, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "HISTORY"}, 1, 20)
	if err != nil || len(history) != 1 || !history[0].Suppressed || history[0].SuppressedUntil == nil {
		t.Fatalf("history after suppress = %+v, err = %v", history, err)
	}
	if stats, err := events.Stats(ctx, ns); err != nil || stats.Suppressed != 1 || stats.Firing != 0 {
		t.Fatalf("stats after suppress = %+v, err = %v", stats, err)
	}

	// 到期时间已过：实时列表重新包含这条事件，历史列表不再计入。
	expired := time.Now().UTC().Add(-time.Minute)
	if err := events.Suppress(ctx, ns, firing[0].UID, &expired); err != nil {
		t.Fatal(err)
	}
	if realtime, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20); err != nil || len(realtime) != 1 {
		t.Fatalf("realtime after expiry = %+v, err = %v", realtime, err)
	}
	if history, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "HISTORY"}, 1, 20); err != nil || len(history) != 0 {
		t.Fatalf("history after expiry = %+v, err = %v", history, err)
	}
	if stats, err := events.Stats(ctx, ns); err != nil || stats.Suppressed != 0 || stats.Firing != 1 {
		t.Fatalf("stats after expiry = %+v, err = %v", stats, err)
	}

	// 再次命中会清掉过期的抑制状态。
	if err := events.ApplyReport(ctx, rule, report); err != nil {
		t.Fatal(err)
	}
	realtime, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil || len(realtime) != 1 || realtime[0].Suppressed {
		t.Fatalf("realtime after re-hit = %+v, err = %v", realtime, err)
	}

	// 永久抑制传空到期时间：持续离开实时列表，历史与统计按抑制计算。
	if err := events.Suppress(ctx, ns, firing[0].UID, nil); err != nil {
		t.Fatal(err)
	}
	if realtime, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20); err != nil || len(realtime) != 0 {
		t.Fatalf("realtime after permanent suppress = %+v, err = %v", realtime, err)
	}
	history, _, err = events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "HISTORY"}, 1, 20)
	if err != nil || len(history) != 1 || !history[0].Suppressed || history[0].SuppressedUntil != nil {
		t.Fatalf("history after permanent suppress = %+v, err = %v", history, err)
	}
	if stats, err := events.Stats(ctx, ns); err != nil || stats.Suppressed != 1 || stats.Firing != 0 {
		t.Fatalf("stats after permanent suppress = %+v, err = %v", stats, err)
	}
}
