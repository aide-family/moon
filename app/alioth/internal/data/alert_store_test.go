package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrulegroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func TestAlertStore(t *testing.T) {
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
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	var kind string
	if err := store.db.QueryRowContext(ctx, `SELECT c.relkind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relname = 'alert_events'`).Scan(&kind); err != nil {
		t.Fatal(err)
	}
	if kind != "p" {
		t.Fatalf("alert_events relkind = %s", kind)
	}

	ns := "ns-alert-" + time.Now().Format("150405.000")
	sources := NewDatasourceRepo(store, slog.Default())
	rules := NewAlertRuleRepo(store, slog.Default())
	t.Cleanup(func() {
		_, _ = store.client.AlertRuleLevel.Delete().Where(alertrulelevel.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertRule.Delete().Where(alertrule.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertRuleGroup.Delete().Where(alertrulegroup.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.Datasource.Delete().Where(datasource.NamespaceUIDEQ(ns)).Exec(ctx)
	})

	created, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: "http://127.0.0.1:9090", Password: "plain-password", Secret: "plain-secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Password != "plain-password" || created.Secret != "plain-secret" || !created.SecretSet {
		t.Fatalf("stored datasource: %+v", created)
	}
	row, err := store.client.Datasource.Query().Where(datasource.UIDEQ(created.UID)).Only(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if row.PasswordCipher == "plain-password" || row.SecretCipher == "plain-secret" || row.PasswordCipher == "" {
		t.Fatal("password is not stored as ciphertext")
	}
	listed, _, err := sources.ListDatasources(ctx, ns, 1, 20, "", "", perm.DatasourceDrivers())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].Password != "" || listed[0].Secret != "" || !listed[0].SecretSet {
		t.Fatalf("list redacts secrets: %+v", listed)
	}
	_, err = sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "prom", Type: "LOGS", Driver: "LOGS_ELASTICSEARCH", Endpoint: "http://127.0.0.1:9200",
	})
	if !merr.IsConflict(err) {
		t.Fatalf("duplicate name: %v", err)
	}
	// 同一地址即使换个名称或指标驱动也属于同一连接，不能再建。
	_, err = sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "prom-copy", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: " http://127.0.0.1:9090/ ",
	})
	if !merr.IsConflict(err) {
		t.Fatalf("duplicate metrics connection: %v", err)
	}
	_, err = sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "prom-vm", Type: "METRICS", Driver: "METRICS_VICTORIA_METRICS",
		Endpoint: "http://127.0.0.1:9090",
	})
	if !merr.IsConflict(err) {
		t.Fatalf("duplicate endpoint across metrics drivers: %v", err)
	}
	// 同地址不同类型是不同的数据面，允许同时存在。
	if _, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "same-host-logs", Type: "LOGS", Driver: "LOGS_LOKI",
		Endpoint: "http://127.0.0.1:9090",
	}); err != nil {
		t.Fatalf("logs datasource on a metrics host: %v", err)
	}
	// 停用后仍占用连接，避免用户用“停用再新增”绕过限制。
	if _, err := sources.UpdateDatasourceStatus(ctx, ns, created.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	_, err = sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "prom-after-disable", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: "http://127.0.0.1:9090",
	})
	if !merr.IsConflict(err) {
		t.Fatalf("disabled datasource should keep its connection: %v", err)
	}
	loki, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "loki", Type: "LOGS", Driver: "LOGS_LOKI",
		Endpoint: "http://127.0.0.1:3100", Secret: "tenant-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if loki.Driver != "LOGS_LOKI" || !loki.SecretSet || loki.Secret != "tenant-token" {
		t.Fatalf("loki datasource: %+v", loki)
	}

	group, err := rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: "core"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: "core"})
	if !merr.IsConflict(err) {
		t.Fatalf("duplicate group: %v", err)
	}
	enabled, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: group.UID, Name: "cpu", Driver: "METRICS_PROMETHEUS", Expr: "up",
		Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "1 + 1", Description: "剩余容量"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 变量写入后要能原样读回，字段和顺序都不能变。
	if len(enabled.Variables) != 1 || enabled.Variables[0] != (bo.AlertRuleVariable{Name: "disk_free", Expr: "1 + 1", Description: "剩余容量"}) {
		t.Fatalf("created variables = %+v", enabled.Variables)
	}
	updated, err := rules.UpdateAlertRule(ctx, bo.AlertRule{
		UID: enabled.UID, NamespaceUID: ns, GroupUID: group.UID, Name: "cpu", Driver: "METRICS_PROMETHEUS", Expr: "up",
		Variables: []bo.AlertRuleVariable{{Name: "load", Expr: "2 + 2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 保存时整列替换：旧变量不能残留，空说明也按空字符串读回。
	if len(updated.Variables) != 1 || updated.Variables[0] != (bo.AlertRuleVariable{Name: "load", Expr: "2 + 2"}) {
		t.Fatalf("updated variables = %+v", updated.Variables)
	}
	disabled, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: group.UID, Name: "mem", Driver: "METRICS_PROMETHEUS", Expr: "up",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rules.UpdateAlertRuleStatus(ctx, ns, disabled.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	logs, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: group.UID, Name: "errors", Driver: "LOGS_ELASTICSEARCH", Expr: "error",
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := rules.ListEvalCandidates(ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, item := range candidates {
		if item.NamespaceUID == ns {
			seen[item.UID] = true
		}
	}
	if !seen[enabled.UID] || seen[disabled.UID] || !seen[logs.UID] {
		t.Fatalf("candidates = %v", seen)
	}
}

func TestDatasourceMatcherIncludesLevel(t *testing.T) {
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
	ns := "ns-matcher-" + time.Now().Format("150405.000")
	t.Cleanup(func() {
		_, _ = store.client.Datasource.Delete().Where(datasource.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertDictionary.Delete().Where(alertdictionary.NamespaceUIDEQ(ns)).Exec(ctx)
	})
	dictionaries := NewAlertDictionaryRepo(store, slog.Default())
	level, err := dictionaries.CreateAlertDictionary(ctx, bo.AlertDictionary{
		NamespaceUID: ns, Name: "低优", Type: "DATASOURCE", Status: "ACTIVE",
	})
	if err != nil {
		t.Fatal(err)
	}
	sources := NewDatasourceRepo(store, slog.Default())
	if _, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "victoriametrics", Type: "METRICS", Driver: "METRICS_VICTORIA_METRICS",
		Endpoint: "http://127.0.0.1:8428", LevelUID: level.UID, Labels: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		matchers string
		want     int
	}{
		{`{"datasource_level":"低优"}`, 1},
		{`{"datasource_level":"高优"}`, 0},
		{`{"datasource":"victoriametrics"}`, 1},
		{`{"datasource_level":"低.+","datasource":"victoria.*"}`, 1},
	}
	for _, tc := range cases {
		items, err := sources.ListDatasourcesByMatchers(ctx, ns, "METRICS_VICTORIA_METRICS", tc.matchers)
		if err != nil {
			t.Fatalf("%s: %v", tc.matchers, err)
		}
		if len(items) != tc.want {
			t.Fatalf("%s: got %d items", tc.matchers, len(items))
		}
	}
}
