package eval

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// failingQuery 模拟数据源请求失败。
type failingQuery struct{ err error }

func (q failingQuery) QueryRange(context.Context, bo.RuleSnapshot, int64, int64, int64) ([]bo.Sample, error) {
	return nil, q.err
}

func TestUpstreamFailureReportsFailedBatch(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: failingQuery{err: merr.Server("upstream query failed")}, log: slog.Default(), now: time.Now}
	slot := time.Now().UTC().Truncate(time.Second)
	err := worker.finish(context.Background(), fakeHold{}, bo.RuleSnapshot{
		UID: "rule-1", DatasourceUID: "ds-1", Driver: "METRICS_PROMETHEUS", EvalSlot: slot,
	})
	if err != nil {
		t.Fatalf("failed eval returned %v", err)
	}
	if len(gateway.reports) != 1 {
		t.Fatalf("reports = %+v", gateway.reports)
	}
	report := gateway.reports[0]
	if !report.Failed || report.Skipped || len(report.Hits) != 0 || report.FailureMessage == "" {
		t.Fatalf("failed report = %+v", report)
	}
	if !report.EvalSlot.Equal(slot) || report.DatasourceUID != "ds-1" {
		t.Fatalf("failed report identity = %+v", report)
	}
}

func TestInvalidExpressionDoesNotReport(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: failingQuery{err: merr.Invalid("invalid query expression")}, log: slog.Default(), now: time.Now}
	err := worker.finish(context.Background(), fakeHold{}, bo.RuleSnapshot{
		UID: "rule-1", Driver: "METRICS_PROMETHEUS", EvalSlot: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("invalid expression returned success")
	}
	if len(gateway.reports) != 0 {
		t.Fatalf("invalid expression reported a datasource failure: %+v", gateway.reports)
	}
}

func TestCanceledRoundDoesNotReport(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: failingQuery{err: merr.Server("upstream query failed")}, log: slog.Default(), now: time.Now}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := worker.finish(ctx, fakeHold{}, bo.RuleSnapshot{
		UID: "rule-1", Driver: "METRICS_PROMETHEUS", EvalSlot: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("canceled round returned success")
	}
	if len(gateway.reports) != 0 {
		t.Fatalf("canceled round reported a datasource failure: %+v", gateway.reports)
	}
}
