package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestQueryVlogsLogsUsesLimitForNewest(t *testing.T) {
	var (
		gotPath, gotQuery, gotStart, gotEnd, gotLimit string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		values := r.URL.Query()
		gotQuery = values.Get("query")
		gotStart = values.Get("start")
		gotEnd = values.Get("end")
		gotLimit = values.Get("limit")
		_, _ = w.Write([]byte("{\"_msg\":\"error\",\"_time\":\"2026-10-20T10:00:00Z\"}\n"))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	body, err := repo.QueryLogs(context.Background(), bo.Datasource{
		UID: "ds-vlogs", Driver: driverLogsVlogs, Endpoint: srv.URL,
	}, "error", 100, 200, 50, "backward")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != vlogsQueryPath {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "error" {
		t.Fatalf("query = %q", gotQuery)
	}
	if gotStart != "100" || gotEnd != "200" || gotLimit != "50" {
		t.Fatalf("start/end/limit = %q/%q/%q", gotStart, gotEnd, gotLimit)
	}
	if body == "" {
		t.Fatal("body is empty")
	}
}

func TestQueryVlogsLogsSortsForOldest(t *testing.T) {
	var gotPath, gotQuery, gotLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		gotLimit = r.URL.Query().Get("limit")
		_, _ = w.Write([]byte(""))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	if _, err := repo.QueryLogs(context.Background(), bo.Datasource{
		UID: "ds-vlogs", Driver: driverLogsVlogs, Endpoint: srv.URL + "/logs",
	}, "error", 100, 200, 20, "forward"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/logs/select/logsql/query" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "error | sort by (_time) | limit 20" {
		t.Fatalf("query = %q", gotQuery)
	}
	if gotLimit != "" {
		t.Fatalf("limit should be part of the LogsQL pipe, got %q", gotLimit)
	}
}

func TestVlogsMetricRangeQueryUsesStatsRange(t *testing.T) {
	var gotPath, gotQuery, gotStart, gotEnd, gotStep, gotLimit, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		values := r.URL.Query()
		gotQuery = values.Get("query")
		gotStart = values.Get("start")
		gotEnd = values.Get("end")
		gotStep = values.Get("step")
		gotLimit = values.Get("limit")
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
	}))
	defer srv.Close()

	expr := `level:error | stats by (app, level) count() as value`
	body, err := newLogTestRepo().RangeQuery(context.Background(), bo.Datasource{
		UID: "ds-vlogs", Driver: driverLogsVlogs, Endpoint: srv.URL + "/logs", Secret: "vlogs-token",
	}, expr, 100, 200, 15)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/logs/select/logsql/stats_query_range" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != expr || gotStart != "100" || gotEnd != "200" || gotStep != "15s" {
		t.Fatalf("query/start/end/step = %q/%q/%q/%q", gotQuery, gotStart, gotEnd, gotStep)
	}
	if gotLimit != "" {
		t.Fatalf("stats range must not send limit, got %q", gotLimit)
	}
	if gotAuth != "Bearer vlogs-token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if body == "" {
		t.Fatal("body is empty")
	}
}
