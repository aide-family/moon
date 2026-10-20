package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestListVictoriaMetricsMetricsUsesLabelValues(t *testing.T) {
	var (
		gotPath       string
		gotStart      int64
		gotEnd        int64
		gotPrometheus bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Path == "/api/v1/metadata" {
			gotPrometheus = true
		}
		gotStart, _ = strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
		gotEnd, _ = strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)
		_, _ = w.Write([]byte(`{"status":"success","data":["up","node_cpu_seconds_total","up",""]}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	metrics, err := repo.ListMetrics(context.Background(), bo.Datasource{
		UID: "ds-vm", Driver: driverMetricsVictoriaMetrics, Endpoint: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPrometheus {
		t.Fatal("victoriametrics must not read the empty prometheus metadata endpoint")
	}
	if gotPath != "/api/v1/label/__name__/values" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotEnd-gotStart != int64(vmMetricLookback/time.Second) {
		t.Fatalf("window = %d..%d", gotStart, gotEnd)
	}
	if len(metrics) != 2 || metrics[0].Name != "node_cpu_seconds_total" || metrics[1].Name != "up" {
		t.Fatalf("metrics = %+v", metrics)
	}
	if metrics[0].Help != "" || metrics[0].Type != "" || metrics[0].Unit != "" {
		t.Fatalf("victoriametrics metrics should not invent metadata: %+v", metrics[0])
	}
}

func TestVictoriaMetricsMetricDetailUsesSeries(t *testing.T) {
	var gotMatch, gotLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/series" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		gotMatch = r.URL.Query().Get("match[]")
		gotLimit = r.URL.Query().Get("limit")
		_, _ = w.Write([]byte(`{"status":"success","data":[{"__name__":"node_cpu_seconds_total","cpu":"0"}]}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	body, err := repo.MetricDetail(context.Background(), bo.Datasource{
		UID: "ds-vm", Driver: driverMetricsVictoriaMetrics, Endpoint: srv.URL,
	}, "node_cpu_seconds_total")
	if err != nil {
		t.Fatal(err)
	}
	if gotMatch != `{__name__="node_cpu_seconds_total"}` {
		t.Fatalf("match[] = %q", gotMatch)
	}
	if gotLimit != strconv.Itoa(vmMetricSeriesLimit) {
		t.Fatalf("limit = %q", gotLimit)
	}
	if body == "" {
		t.Fatal("body is empty")
	}
}

func TestVictoriaMetricsMetricDetailRejectsInvalidName(t *testing.T) {
	repo := newLogTestRepo()
	for _, name := range []string{"", `bad"name`, "1starts_with_digit", "metric-name", "指标"} {
		if _, err := repo.MetricDetail(context.Background(), bo.Datasource{
			UID: "ds-vm", Driver: driverMetricsVictoriaMetrics, Endpoint: "http://127.0.0.1:1",
		}, name); !merr.IsInvalid(err) {
			t.Fatalf("MetricDetail(%q) error = %v", name, err)
		}
	}
}

func TestVictoriaMetricsQueryAndStatusKeepEndpointPrefix(t *testing.T) {
	var (
		gotPath  string
		gotQuery string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		if r.URL.Path == "/select/0/prometheus/api/v1/query_range" {
			_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	item := bo.Datasource{
		UID: "ds-vm", Driver: driverMetricsVictoriaMetrics,
		Endpoint: srv.URL + "/select/0/prometheus",
	}
	if _, err := repo.DatasourceStatus(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/select/0/prometheus/api/v1/status/buildinfo" {
		t.Fatalf("status path = %q", gotPath)
	}
	if _, err := repo.InstantQuery(context.Background(), item, "up", 0); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/select/0/prometheus/api/v1/query" || gotQuery != "up" {
		t.Fatalf("instant path = %q query = %q", gotPath, gotQuery)
	}
	if _, err := repo.RangeQuery(context.Background(), item, "rate(foo[5m])", 100, 200, 15); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/select/0/prometheus/api/v1/query_range" {
		t.Fatalf("range path = %q", gotPath)
	}
	if gotQuery != "rate(foo[5m])" {
		t.Fatalf("range query = %q", gotQuery)
	}
}

func TestListMetricsDispatchesByDriver(t *testing.T) {
	repo := newLogTestRepo()
	if _, err := repo.ListMetrics(context.Background(), bo.Datasource{
		UID: "ds-loki", Driver: driverLogsLoki, Endpoint: "http://127.0.0.1:1",
	}); !merr.IsInvalid(err) {
		t.Fatalf("logs driver error = %v", err)
	}
	if _, err := repo.InstantQuery(context.Background(), bo.Datasource{
		UID: "ds-loki", Driver: driverLogsLoki, Endpoint: "http://127.0.0.1:1",
	}, "up", 0); !merr.IsInvalid(err) {
		t.Fatalf("logs instant query error = %v", err)
	}
}

func TestValidMetricName(t *testing.T) {
	for _, name := range []string{"up", "_internal", ":recorded", "node_cpu_seconds_total", "a1_b2"} {
		if !validMetricName(name) {
			t.Fatalf("validMetricName(%q) = false", name)
		}
	}
	for _, name := range []string{"", "1up", "metric-name", `bad"name`, "指标"} {
		if validMetricName(name) {
			t.Fatalf("validMetricName(%q) = true", name)
		}
	}
}
