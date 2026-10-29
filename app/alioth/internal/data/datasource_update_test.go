package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/biz/datasource"
	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	dsent "github.com/aide-family/moon/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

type testNamespaces struct{ repository.NamespaceRepo }

func (testNamespaces) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	return &bo.Namespace{UID: uid, Name: uid, Status: "ACTIVE"}, nil
}

type testRoles struct{ repository.RoleRepo }

func (testRoles) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return map[string]struct{}{
		perm.DatasourceWrite:                           {},
		perm.DatasourceRead:                            {},
		perm.MenuPermission(perm.CapabilityDatasource): {},
	}, nil
}

type testFeatures struct{ repository.FeatureRepo }

func (testFeatures) EffectiveFeatures(context.Context, string) ([]string, error) {
	return perm.FeatureCodes(), nil
}

func (testFeatures) EffectiveModules(context.Context, string) ([]string, error) {
	return perm.FeatureCodes(), nil
}

func (testFeatures) EffectiveCapabilities(context.Context, string) ([]string, error) {
	items := perm.Capabilities()
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	return codes, nil
}

// TestUpdateDatasourceLevelRecoversEvents 校验修改数据源等级后，旧的 FIRING 事件
// 会先收口到历史库，下一次评测才能按新等级生成新事件，而不是一直显示旧等级。
func TestUpdateDatasourceLevelRecoversEvents(t *testing.T) {
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
	ns := "ns-ds-update-" + time.Now().Format("150405.000")
	userUID := "user-" + ns
	dsUID := "ds-" + ns
	t.Cleanup(func() {
		_, _ = store.client.Datasource.Delete().Where(dsent.UIDEQ(dsUID)).Exec(ctx)
		_, _ = store.client.AlertDictionary.Delete().Where(alertdictionary.NamespaceUIDEQ(ns)).Exec(ctx)
		_, _ = store.client.AlertEvent.Delete().Where(alertevent.NamespaceUIDEQ(ns)).Exec(ctx)
		_ = store.history.dropTenant(context.Background(), ns)
	})

	dictionaries := NewAlertDictionaryRepo(store, slog.Default())
	for _, name := range []string{"P1", "P2"} {
		if _, err := dictionaries.CreateAlertDictionary(ctx, bo.AlertDictionary{
			NamespaceUID: ns, Name: name, Type: "DATASOURCE", Status: "ACTIVE",
		}); err != nil {
			t.Fatal(err)
		}
	}
	levels, _, err := dictionaries.ListAlertDictionaries(ctx, ns, 1, 10, "DATASOURCE", "")
	if err != nil || len(levels) != 2 {
		t.Fatalf("levels = %+v, err = %v", levels, err)
	}

	sources := NewDatasourceRepo(store, slog.Default())
	created, err := sources.CreateDatasource(ctx, bo.Datasource{
		NamespaceUID: ns, Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: "http://127.0.0.1:9090", LevelUID: levels[0].UID, Labels: "{}",
	})
	if err != nil {
		t.Fatal(err)
	}
	events := NewAlertEventRepo(store, slog.Default())
	rule := bo.AlertRule{UID: "rule-" + ns, NamespaceUID: ns, DatasourceUID: created.UID, Expr: "up", Summary: "cpu"}
	if err := events.ApplyReport(ctx, rule, bo.EvalReport{
		RuleUID: rule.UID, DatasourceUID: created.UID, EvalSlot: time.Now().UTC().Truncate(time.Second),
		Hits: []bo.EvalHit{{Fingerprint: "fp-1", Labels: `{"job":"api"}`, Value: "1"}},
	}); err != nil {
		t.Fatal(err)
	}

	notifications := NewNotificationRepo(store, slog.Default())
	gateway := datasource.NewGateway(sources, dictionaries, events, notifications, access.Checker{
		Namespaces: testNamespaces{}, Roles: testRoles{}, Features: testFeatures{}, Log: slog.Default(),
	})
	runCtx := actor.WithUser(actor.WithNamespace(ctx, ns), userUID)
	if _, err := gateway.UpdateDatasource(runCtx, bo.Datasource{
		UID: created.UID, NamespaceUID: ns, Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS",
		Endpoint: "http://127.0.0.1:9090", LevelUID: levels[1].UID, Labels: "{}",
	}); err != nil {
		t.Fatal(err)
	}
	realtime, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "REALTIME"}, 1, 20)
	if err != nil || len(realtime) != 0 {
		t.Fatalf("realtime after level change = %+v, err = %v", realtime, err)
	}
	history, _, err := events.ListEvents(ctx, ns, bo.AlertEventListFilter{View: "HISTORY"}, 1, 20)
	if err != nil || len(history) != 1 {
		t.Fatalf("history after level change = %+v, err = %v", history, err)
	}
}
