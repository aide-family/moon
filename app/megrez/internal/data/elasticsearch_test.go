package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
)

func TestElasticsearchQueryRangeCountsDocumentsPerBucket(t *testing.T) {
	var (
		gotMethod, gotPath, gotAuth string
		gotBody                     map[string]any
	)
	startAt := time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)
	endAt := startAt.Add(30 * time.Second)
	body := fmt.Sprintf(`{"aggregations":{"buckets":{"buckets":[
{"key":%d,"doc_count":0},
{"key":%d,"doc_count":7}]}}}`,
		startAt.UnixMilli(), startAt.Add(15*time.Second).UnixMilli())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(payload, &gotBody); err != nil {
			t.Fatalf("request body = %s", payload)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	samples, err := rangeQuery{}.QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "es-rule", Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL,
		Expr: `{"match":{"level":"error"}}`, IndexPattern: "alioth-smoke*", TimeField: "@timestamp",
		Secret: "es-token",
	}, startAt.Unix(), endAt.Unix(), 15)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q，date_histogram 需要 POST", gotMethod)
	}
	if gotPath != "/alioth-smoke*/_search" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer es-token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	// 只要聚合结果，不拉命中文档。
	if gotBody["size"] != float64(0) {
		t.Fatalf("size = %v", gotBody["size"])
	}
	histogram := gotBody["aggs"].(map[string]any)["buckets"].(map[string]any)["date_histogram"].(map[string]any)
	if histogram["fixed_interval"] != "15s" || histogram["field"] != "@timestamp" {
		t.Fatalf("date_histogram = %v", histogram)
	}
	if len(samples) != 1 {
		t.Fatalf("samples = %+v", samples)
	}
	sample := samples[0]
	if sample.Step != 15 || len(sample.Points) != 2 {
		t.Fatalf("sample = %+v", sample)
	}
	if sample.Points[0].Value != 0 || sample.Points[0].Raw != "0" {
		t.Fatalf("points[0] = %+v", sample.Points[0])
	}
	if sample.Points[1].Value != 7 || sample.Points[1].Raw != "7" {
		t.Fatalf("points[1] = %+v", sample.Points[1])
	}
	if !sample.Points[1].At.Equal(startAt.Add(15 * time.Second)) {
		t.Fatalf("points[1].At = %s", sample.Points[1].At)
	}
}

func TestElasticsearchQueryRangeUsesBasicAuth(t *testing.T) {
	var gotUser, gotPassword string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPassword, _ = r.BasicAuth()
		_, _ = w.Write([]byte(`{"aggregations":{"buckets":{"buckets":[{"key":1000,"doc_count":1}]}}}`))
	}))
	defer srv.Close()

	if _, err := (rangeQuery{}).QueryRange(context.Background(), bo.RuleSnapshot{
		Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL, Expr: "level:error",
		Username: "elastic", Password: "secret",
	}, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	if gotUser != "elastic" || gotPassword != "secret" {
		t.Fatalf("basic auth = %q/%q", gotUser, gotPassword)
	}
}

func TestElasticsearchQueryRangeRejectsBadResponses(t *testing.T) {
	for _, payload := range []string{
		`{"error":{"type":"index_not_found_exception"}}`,
		`{"aggregations":{}}`,
		`not json`,
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(payload))
		}))
		_, err := (rangeQuery{}).QueryRange(context.Background(), bo.RuleSnapshot{
			Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL, Expr: "level:error",
		}, 0, 15, 15)
		srv.Close()
		if err == nil {
			t.Fatalf("payload %q should fail：查询失败不能被当成零命中", payload)
		}
	}
}

func TestElasticsearchQueryRangeRejectsEmptyExpression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("空表达式不应发出上游请求")
	}))
	defer srv.Close()

	_, err := (rangeQuery{}).QueryRange(context.Background(), bo.RuleSnapshot{
		Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL, Expr: "  ",
	}, 0, 15, 15)
	if err == nil {
		t.Fatal("empty expression should fail")
	}
}
