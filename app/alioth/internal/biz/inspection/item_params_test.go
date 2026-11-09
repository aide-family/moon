package inspection

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestNormalizeItemParamsMetrics(t *testing.T) {
	raw, err := normalizeItemParams("METRICS_QUERY", `{"unit":"s","thresholds":[{"op":">","value":0.8,"status":"warn"}]}`)
	if err != nil {
		t.Fatalf("normalize metrics: %v", err)
	}
	params := parseMetricsParams(raw)
	if params.Unit != "s" || len(params.Thresholds) != 1 || params.Thresholds[0].Status != "WARN" {
		t.Fatalf("unexpected params: %+v", params)
	}
	if _, err := normalizeItemParams("METRICS_QUERY", `{"unknown":true}`); err == nil {
		t.Fatal("unknown key should be rejected")
	}
	if _, err := normalizeItemParams("METRICS_QUERY", `{"thresholds":[{"op":"~","value":1,"status":"WARN"}]}`); err == nil {
		t.Fatal("invalid operator should be rejected")
	}
}

func TestNormalizeItemParamsLogs(t *testing.T) {
	raw, err := normalizeItemParams("LOGS_QUERY", `{"levels":["error","WARN"],"limit":100}`)
	if err != nil {
		t.Fatalf("normalize logs: %v", err)
	}
	params := parseLogsParams(raw)
	if params.Limit != 100 || params.Direction != "BACKWARD" || strings.Join(params.Levels, ",") != "ERROR,WARN" {
		t.Fatalf("unexpected logs params: %+v", params)
	}
	if _, err := normalizeItemParams("LOGS_QUERY", `{"limit":10}`); err == nil {
		t.Fatal("limit below range should be rejected")
	}
}

func TestNormalizeItemParamsProbe(t *testing.T) {
	raw, err := normalizeItemParams("PROBE_TARGET", `{"group_uids":["g1"],"metric":"p95"}`)
	if err != nil {
		t.Fatalf("normalize probe: %v", err)
	}
	params := parseProbeParams(raw)
	if params.Metric != "P95" || params.LatencyAboveMS != 500 || params.WarnOnDown == nil || !*params.WarnOnDown {
		t.Fatalf("unexpected probe params: %+v", params)
	}
	if _, err := normalizeItemParams("PROBE_TARGET", `{"group_uids":[]}`); err == nil {
		t.Fatal("probe without group should be rejected")
	}
}

func TestNormalizeItemParamsAlertLabels(t *testing.T) {
	raw, err := normalizeItemParams("ALERT_EVENTS", `{"labels":"service=order, env=prod"}`)
	if err != nil {
		t.Fatalf("normalize alert: %v", err)
	}
	var params alertParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	var labels map[string]string
	if err := json.Unmarshal([]byte(params.Labels), &labels); err != nil {
		t.Fatal(err)
	}
	if labels["service"] != "order" || labels["env"] != "prod" {
		t.Fatalf("unexpected labels: %s", params.Labels)
	}
	if _, err := normalizeItemParams("NOTE", `{"x":1}`); err == nil {
		t.Fatal("note params should be rejected")
	}
}

func TestMetricThresholdStatus(t *testing.T) {
	series := []bo.InspectionSeries{{Points: []bo.InspectionPoint{{Value: 0.9}}}}
	if got := metricThresholdStatus(series, []metricThreshold{{Op: ">", Value: 0.8, Status: "WARN"}}); got != "WARN" {
		t.Fatalf("want WARN, got %q", got)
	}
	if got := metricThresholdStatus(series, []metricThreshold{{Op: "<", Value: 1, Status: "FAILED"}}); got != "FAILED" {
		t.Fatalf("want FAILED, got %q", got)
	}
	if got := metricThresholdStatus(series, []metricThreshold{{Op: ">", Value: 1, Status: "WARN"}}); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestProbeStatusByMetric(t *testing.T) {
	warnOnDown := true
	base := probeParams{Metric: "AVAILABILITY", AvailabilityBelow: 99.5, WarnOnDown: &warnOnDown}
	if got := probeStatus(base, probeSummary{checks: 10, ok: 10}, 100); got != "正常" {
		t.Fatalf("availability normal: %s", got)
	}
	if got := probeStatus(base, probeSummary{checks: 10, ok: 9}, 90); got != "异常" {
		t.Fatalf("availability abnormal: %s", got)
	}
	latency := probeParams{Metric: "LATENCY", LatencyAboveMS: 500}
	if got := probeStatus(latency, probeSummary{checks: 2, avgLatency: 800}, 100); got != "异常" {
		t.Fatalf("latency abnormal: %s", got)
	}
	cert := probeParams{Metric: "CERT_DAYS", CertDaysBelow: 14}
	if got := probeStatus(cert, probeSummary{checks: 2, certDays: 7}, 100); got != "异常" {
		t.Fatalf("cert abnormal: %s", got)
	}
}

func TestParseLogSamplesFormats(t *testing.T) {
	es := `{"hits":{"total":{"value":2},"hits":[{"_source":{"message":"ERROR upstream timeout","@timestamp":"2026-11-09T10:00:00Z"}}]}}`
	logs, total := parseLogSamples(es)
	if total != 2 || len(logs) != 1 || logs[0].Level != "ERROR" {
		t.Fatalf("unexpected es logs: %+v total=%d", logs, total)
	}
	ndjson := "{\"_msg\":\"WARN slow request\",\"_time\":\"2026-11-09T10:00:00Z\"}\n{\"_msg\":\"INFO ok\"}"
	logs, total = parseLogSamples(ndjson)
	if total != 2 || len(logs) != 2 || logs[0].Level != "WARN" {
		t.Fatalf("unexpected ndjson logs: %+v total=%d", logs, total)
	}
}
