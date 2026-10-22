package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
)

func TestLokiQueryRangeUsesLogQLEndpointAndDecodesSeries(t *testing.T) {
	var gotPath, gotQuery, gotStart, gotEnd, gotStep, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		gotStart = r.URL.Query().Get("start")
		gotEnd = r.URL.Query().Get("end")
		gotStep = r.URL.Query().Get("step")
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"app":"api","level":"error"},"values":[[100,"2"],[115,"4"]]}]}}`))
	}))
	defer srv.Close()

	query := rangeQuery{}
	samples, err := query.QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "loki-rule", Driver: "LOGS_LOKI", Endpoint: srv.URL + "/tenant-a",
		Expr: `sum by (app) (count_over_time({app="api"} |= "error" [5m]))`, Secret: "loki-token",
	}, 100, 200, 15)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/tenant-a/loki/api/v1/query_range" {
		t.Fatalf("path = %q", gotPath)
	}
	wantStart := strconv.FormatInt(time.Unix(100, 0).UnixNano(), 10)
	wantEnd := strconv.FormatInt(time.Unix(200, 0).UnixNano(), 10)
	if gotQuery != `sum by (app) (count_over_time({app="api"} |= "error" [5m]))` || gotStart != wantStart || gotEnd != wantEnd || gotStep != "15s" {
		t.Fatalf("query/start/end/step = %q/%q/%q/%q, want %q/%q/%q/%q", gotQuery, gotStart, gotEnd, gotStep, `sum by (app) (count_over_time({app="api"} |= "error" [5m]))`, wantStart, wantEnd, "15s")
	}
	if gotAuth != "Bearer loki-token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if len(samples) != 1 || samples[0].Labels["app"] != "api" || samples[0].Labels["level"] != "error" || len(samples[0].Points) != 2 {
		t.Fatalf("samples = %+v", samples)
	}
	if samples[0].Points[1].Value != 4 || samples[0].Points[1].At.Unix() != 115 {
		t.Fatalf("point = %+v", samples[0].Points[1])
	}
}

func TestLokiInstantQueryUsesLokiAPI(t *testing.T) {
	var gotPath, gotQuery, gotAt string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		gotAt = r.URL.Query().Get("time")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"app":"api"},"value":[123.456,"9"]}]}}`))
	}))
	defer srv.Close()

	at := time.Unix(123, 456000000).UTC()
	samples, err := (rangeQuery{}).QueryInstant(context.Background(), bo.RuleSnapshot{
		UID: "loki-rule", Driver: "LOGS_LOKI", Endpoint: srv.URL,
	}, `sum(count_over_time({app="api"}[5m]))`, at)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != lokiQueryPath || gotQuery != `sum(count_over_time({app="api"}[5m]))` || gotAt != "123456000000" {
		t.Fatalf("path/query/time = %q/%q/%q", gotPath, gotQuery, gotAt)
	}
	if len(samples) != 1 || samples[0].Labels["app"] != "api" || samples[0].Points[0].Value != 9 {
		t.Fatalf("samples = %+v", samples)
	}
}

func TestLokiRangeAcceptsEmptyNumericMatrix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
	}))
	defer srv.Close()

	samples, err := (rangeQuery{}).QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "loki-rule", Driver: "LOGS_LOKI", Endpoint: srv.URL, Expr: `sum(count_over_time({app="api"}[5m]))`,
	}, 100, 200, 15)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 0 {
		t.Fatalf("empty matrix samples = %+v", samples)
	}
}

func TestLokiRangeRejectsRawLogStreamResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"streams","result":[]}}`))
	}))
	defer srv.Close()

	_, err := (rangeQuery{}).QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "loki-rule", Driver: "LOGS_LOKI", Endpoint: srv.URL, Expr: `{app="api"} |= "error"`,
	}, 100, 200, 15)
	if err == nil {
		t.Fatal("raw log stream response must not be treated as an empty metric result")
	}
}
