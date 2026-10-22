package data

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
)

func TestVlogsQueryRangeUsesStatsRangeAndDecodesMatrix(t *testing.T) {
	var gotPath, gotQuery, gotStart, gotEnd, gotStep, gotLimit, gotAuth string
	startAt := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	next := startAt.Add(15 * time.Second)
	body := fmt.Sprintf(`{"status":"success","data":{"resultType":"matrix","result":[
{"metric":{"__name__":"value","app":"api","level":"error"},"values":[[%d,"2"],[%d,"4"]]},
{"metric":{"__name__":"value","app":"web","level":"error"},"values":[[%d,"1"]]}]}}`,
		startAt.Unix(), next.Unix(), startAt.Unix())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		values := r.URL.Query()
		gotQuery = values.Get("query")
		gotStart = values.Get("start")
		gotEnd = values.Get("end")
		gotStep = values.Get("step")
		gotLimit = values.Get("limit")
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	expr := `level:error | stats by (app, level) count() as value`
	endAt := startAt.Add(30 * time.Second)
	samples, err := rangeQuery{}.QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "vlogs-rule", Driver: "LOGS_VLOGS", Endpoint: srv.URL + "/logs", Expr: expr, Secret: "vlogs-token",
	}, startAt.Unix(), endAt.Unix(), 15)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/logs/select/logsql/stats_query_range" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != expr || gotStart != strconv.FormatInt(startAt.Unix(), 10) || gotEnd != strconv.FormatInt(endAt.Unix(), 10) {
		t.Fatalf("query/start/end = %q/%q/%q", gotQuery, gotStart, gotEnd)
	}
	if gotStep != "15s" || gotLimit != "" {
		t.Fatalf("step/limit = %q/%q", gotStep, gotLimit)
	}
	if gotAuth != "Bearer vlogs-token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if len(samples) != 2 {
		t.Fatalf("samples = %+v", samples)
	}
	api := samples[0]
	web := samples[1]
	if api.Labels["app"] != "api" || api.Labels["__name__"] != "value" || len(api.Points) != 2 || api.Step != 15 {
		t.Fatalf("api sample = %+v", api)
	}
	if api.Points[0].Value != 2 || api.Points[1].Value != 4 || api.Points[1].Raw != "4" {
		t.Fatalf("api points = %+v", api.Points)
	}
	if web.Labels["app"] != "web" || web.Labels["level"] != "error" || web.Points[0].Value != 1 {
		t.Fatalf("web sample = %+v", web)
	}
}

func TestVlogsQueryInstantUsesStatsQuery(t *testing.T) {
	var gotPath, gotQuery, gotTime, gotStart, gotEnd string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		values := r.URL.Query()
		gotQuery = values.Get("query")
		gotTime = values.Get("time")
		gotStart = values.Get("start")
		gotEnd = values.Get("end")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"value","app":"api"},"value":[123,"9"]}]}}`))
	}))
	defer srv.Close()

	at := time.Unix(123, 0).UTC()
	samples, err := rangeQuery{}.QueryInstant(context.Background(), bo.RuleSnapshot{
		UID: "vlogs-rule", Driver: "LOGS_VLOGS", Endpoint: srv.URL,
	}, `level:error | stats by (app) count() as value`, at)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/select/logsql/stats_query" || gotQuery != `level:error | stats by (app) count() as value` {
		t.Fatalf("path/query = %q/%q", gotPath, gotQuery)
	}
	if gotTime != "123" || gotStart != "108" || gotEnd != "123" {
		t.Fatalf("time/start/end = %q/%q/%q", gotTime, gotStart, gotEnd)
	}
	if len(samples) != 1 || samples[0].Labels["app"] != "api" || samples[0].Points[0].Value != 9 {
		t.Fatalf("samples = %+v", samples)
	}
}

func TestVlogsRangeAcceptsEmptyMatrix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
	}))
	defer srv.Close()

	samples, err := rangeQuery{}.QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "vlogs-rule", Driver: "LOGS_VLOGS", Endpoint: srv.URL, Expr: `error | stats count() as value`,
	}, 100, 200, 15)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 0 {
		t.Fatalf("empty matrix samples = %+v", samples)
	}
}

func TestVlogsRangeRejectsNonMatrixOrNonNumericResponses(t *testing.T) {
	cases := []string{
		`{"_msg":"error: request timeout","_time":"2026-10-20T10:00:00Z"}` + "\n",
		`{"status":"success","data":{"resultType":"vector","result":[]}}`,
		`{"status":"error","error":"bad LogsQL"}`,
		`not-json`,
		`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"value"},"values":[[123,"not-a-number"]]}]}}`,
		`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"value"},"values":[[123,"NaN"]]}]}}`,
		`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"value"},"values":[]}]}}`,
		`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"value"},"values":[[123]]}]}}`,
	}
	for _, body := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(body))
		}))
		_, err := rangeQuery{}.QueryRange(context.Background(), bo.RuleSnapshot{
			UID: "vlogs-rule", Driver: "LOGS_VLOGS", Endpoint: srv.URL, Expr: `error | stats values(level) as value`,
		}, 100, 200, 15)
		srv.Close()
		if err == nil {
			t.Fatalf("body %q should be rejected as a non-numeric matrix response", body)
		}
	}
}
