package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerteventfiring"
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
		Hits: []bo.EvalHit{{Fingerprint: "fp-1", Labels: `{"job":"a"}`, Value: "1", Summary: "cpu"}},
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
