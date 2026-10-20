package eval

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestHitsConditions(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	points := []bo.Point{
		{At: now, Value: 2, Raw: "2"},
		{At: now.Add(15 * time.Second), Value: 3, Raw: "3"},
	}
	rule := bo.RuleSnapshot{UID: "rule-1", DatasourceUID: "ds-1", Summary: "cpu", Levels: []bo.Level{{
		Condition: "GT", Sample: "MAX", Threshold: "1", LevelUID: "high",
	}}}
	hits := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(hits) != 1 || hits[0].Value != "3" || hits[0].LevelUID != "high" || hits[0].Fingerprint == "" {
		t.Fatalf("max hit = %+v", hits)
	}
	rule.Levels[0].Sample = "MIN"
	rule.Levels[0].Threshold = "2"
	if len(Hits(rule, []bo.Sample{{Points: points}})) != 0 {
		t.Fatal("min should miss")
	}
	rule.Levels[0].Sample = "FOR"
	rule.Levels[0].Threshold = "1"
	rule.Levels[0].ForSeconds = 15
	got := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(got) != 1 || got[0].Value != "3" {
		t.Fatalf("for hit = %+v", got)
	}
	rule.Summary = "测试告警, value: {{ $value }} job {{ $labels.job }}"
	rule.Description = "测试告警, value: {{ . }}"
	rule.Levels[0].Condition = "GT"
	rule.Levels[0].Sample = "MAX"
	rule.Levels[0].Threshold = "1"
	filled := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(filled) != 1 || filled[0].Summary != "测试告警, value: 3 job api" || filled[0].Description != "测试告警, value: 3" {
		t.Fatalf("template hit = %+v", filled)
	}
	rule.Summary = "{{"
	if got := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}}); len(got) != 1 || got[0].Summary != "{{" {
		t.Fatalf("broken template = %+v", got)
	}
	rule.Summary = "cpu"
	rule.Description = ""
	rule.Levels[0].Sample = "FOR"
	rule.Levels[0].Condition = "IN"
	rule.Levels[0].Threshold = "3,9"
	if len(Hits(rule, []bo.Sample{{Points: points}})) != 0 {
		t.Fatal("for in should miss when a point is outside")
	}
}

type fakeHold struct{ renew error }

func (h fakeHold) Renew(context.Context, time.Duration) error { return h.renew }
func (h fakeHold) Release(context.Context) error              { return nil }

type fakeGateway struct{ reports []bo.EvalReport }

func (g *fakeGateway) Claim(context.Context) ([]bo.RuleSnapshot, error) { return nil, nil }
func (g *fakeGateway) Report(_ context.Context, report bo.EvalReport) error {
	g.reports = append(g.reports, report)
	return nil
}

type fakeQuery struct{}

func (fakeQuery) QueryRange(context.Context, bo.RuleSnapshot, int64, int64, int64) ([]bo.Sample, error) {
	return []bo.Sample{}, nil
}

func TestRenewFailureDoesNotReport(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: fakeQuery{}, log: slog.Default(), now: time.Now}
	err := worker.finish(context.Background(), fakeHold{renew: merr.Server("lock failed")}, bo.RuleSnapshot{
		UID: "rule-1", Driver: "METRICS_PROMETHEUS", EvalSlot: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("renew failure returned success")
	}
	if len(gateway.reports) != 0 {
		t.Fatalf("reported after renew failure: %+v", gateway.reports)
	}
}

type failLocker struct{}

func (failLocker) Available(context.Context) error { return merr.Server("redis down") }
func (failLocker) Acquire(context.Context, string, time.Duration) (repository.Hold, error) {
	return fakeHold{}, nil
}

func TestClaimFailureStretches(t *testing.T) {
	worker := &Worker{locks: failLocker{}, gateway: &fakeGateway{}, log: slog.Default(), now: time.Now}
	if !worker.once(context.Background()) {
		t.Fatal("redis failure should stretch")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)
	if !worker.once(ctx) {
		t.Fatal("deadline should stretch")
	}
	cancel()
	stopped, stop := context.WithCancel(context.Background())
	stop()
	if worker.once(stopped) {
		t.Fatal("shutdown should not stretch")
	}
}

func TestUnimplementedDriverReportsSkipped(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: fakeQuery{}, log: slog.Default(), now: time.Now}
	slot := time.Now().UTC().Truncate(time.Second)
	for _, driver := range []string{"LOGS_ELASTICSEARCH", "TRACE_JAEGER"} {
		gateway.reports = nil
		if err := worker.finish(context.Background(), fakeHold{}, bo.RuleSnapshot{UID: "rule-1", Driver: driver, EvalSlot: slot}); err != nil {
			t.Fatal(err)
		}
		if len(gateway.reports) != 1 || !gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 0 {
			t.Fatalf("driver %s report = %+v", driver, gateway.reports)
		}
	}
}
