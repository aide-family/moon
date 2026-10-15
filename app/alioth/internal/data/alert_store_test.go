package data

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulegroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestAlertStore(t *testing.T) {
	source := os.Getenv("ALIOTH_TEST_DSN")
	if source == "" {
		source = "postgres://alioth:alioth@127.0.0.1:5432/alioth?sslmode=disable"
	}
	ctx := context.Background()
	store, cleanup, err := NewData(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		log.DefaultLogger,
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
	sources := NewDatasourceRepo(store, log.DefaultLogger)
	rules := NewAlertRuleRepo(store, log.DefaultLogger)
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
	listed, _, err := sources.ListDatasources(ctx, ns, 1, 20, "", "")
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

	group, err := rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: "core"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = rules.CreateAlertRuleGroup(ctx, bo.AlertRuleGroup{NamespaceUID: ns, Name: "core"})
	if !merr.IsConflict(err) {
		t.Fatalf("duplicate group: %v", err)
	}
	enabled, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: group.UID, Name: "cpu", Type: "METRICS", Driver: "METRICS_PROMETHEUS", Expr: "up",
	})
	if err != nil {
		t.Fatal(err)
	}
	disabled, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: group.UID, Name: "mem", Type: "METRICS", Driver: "METRICS_PROMETHEUS", Expr: "up",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rules.UpdateAlertRuleStatus(ctx, ns, disabled.UID, "DISABLED"); err != nil {
		t.Fatal(err)
	}
	logs, err := rules.CreateAlertRule(ctx, bo.AlertRule{
		NamespaceUID: ns, GroupUID: group.UID, Name: "errors", Type: "LOGS", Driver: "LOGS_ELASTICSEARCH", Expr: "error",
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
