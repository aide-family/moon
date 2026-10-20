package data

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func newLogTestRepo() *datasourceRepo {
	return &datasourceRepo{&repoBase{log: slog.Default()}}
}

func TestQueryLogsSendsLokiRangeRequest(t *testing.T) {
	var (
		gotPath, gotQuery, gotStart, gotEnd, gotLimit, gotDirection, gotAuth string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		values := r.URL.Query()
		gotQuery = values.Get("query")
		gotStart = values.Get("start")
		gotEnd = values.Get("end")
		gotLimit = values.Get("limit")
		gotDirection = values.Get("direction")
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"streams","result":[]}}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	body, err := repo.QueryLogs(context.Background(), bo.Datasource{
		UID: "ds-loki", Endpoint: srv.URL, Secret: "token",
	}, `{app="api"} |= "error"`, 100, 200, 50, "forward")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/loki/api/v1/query_range" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != `{app="api"} |= "error"` {
		t.Fatalf("query = %q", gotQuery)
	}
	wantStart := strconv.FormatInt(time.Unix(100, 0).UnixNano(), 10)
	wantEnd := strconv.FormatInt(time.Unix(200, 0).UnixNano(), 10)
	if gotStart != wantStart || gotEnd != wantEnd {
		t.Fatalf("start/end = %q/%q, want %q/%q", gotStart, gotEnd, wantStart, wantEnd)
	}
	if gotLimit != "50" || gotDirection != "forward" {
		t.Fatalf("limit/direction = %q/%q", gotLimit, gotDirection)
	}
	if gotAuth != "Bearer token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if body != `{"status":"success","data":{"resultType":"streams","result":[]}}` {
		t.Fatalf("body = %q", body)
	}
}

func TestDatasourceStatusPathByDriver(t *testing.T) {
	cases := []struct {
		driver string
		path   string
	}{
		{driver: "LOGS_LOKI", path: "/ready"},
		{driver: "METRICS_PROMETHEUS", path: "/api/v1/status/buildinfo"},
		{driver: "METRICS_VICTORIA_METRICS", path: "/api/v1/status/buildinfo"},
	}
	for _, tc := range cases {
		t.Run(tc.driver, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte("ready"))
			}))
			defer srv.Close()

			repo := newLogTestRepo()
			status, err := repo.DatasourceStatus(context.Background(), bo.Datasource{
				UID: "ds-" + tc.driver, Driver: tc.driver, Endpoint: srv.URL,
			})
			if err != nil {
				t.Fatal(err)
			}
			if status != "up" || gotPath != tc.path {
				t.Fatalf("status = %q path = %q", status, gotPath)
			}
		})
	}

	if _, err := statusPath("LOGS_ELASTICSEARCH"); !merr.IsInvalid(err) {
		t.Fatalf("unsupported driver error = %v", err)
	}
}

func TestQueryLogsReportsUpstreamFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	if _, err := repo.QueryLogs(context.Background(), bo.Datasource{UID: "ds-loki", Endpoint: srv.URL}, "{}", 1, 2, 10, "backward"); err == nil {
		t.Fatal("upstream failure should return an error")
	}
	if _, err := repo.DatasourceStatus(context.Background(), bo.Datasource{UID: "ds-loki", Driver: "LOGS_LOKI", Endpoint: srv.URL}); err == nil {
		t.Fatal("not ready datasource should return an error")
	}
}
