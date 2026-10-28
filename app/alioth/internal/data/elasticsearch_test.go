package data

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func TestQueryElasticsearchLogsSendsSearchRequest(t *testing.T) {
	var (
		gotMethod, gotPath, gotAuth string
		gotBody                     map[string]any
	)
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
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":1},"hits":[]}}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	body, err := repo.QueryLogs(context.Background(), bo.Datasource{
		UID: "ds-es", Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL,
		IndexPattern: "logs-*", TimeField: "event.time", Secret: "token",
	}, `{"match":{"level":"error"}}`, 100, 200, 50, "forward")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/logs-*/_search" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	if body != `{"hits":{"total":{"value":1},"hits":[]}}` {
		t.Fatalf("body = %q", body)
	}
	if gotBody["size"] != float64(50) {
		t.Fatalf("size = %v", gotBody["size"])
	}
	sortFields, ok := gotBody["sort"].([]any)
	if !ok || len(sortFields) != 1 {
		t.Fatalf("sort = %v", gotBody["sort"])
	}
	sortField, _ := sortFields[0].(map[string]any)
	if order := sortField["event.time"].(map[string]any)["order"]; order != "asc" {
		t.Fatalf("sort order = %v", order)
	}
	query, _ := gotBody["query"].(map[string]any)
	boolQuery, _ := query["bool"].(map[string]any)
	filters, _ := boolQuery["filter"].([]any)
	rangeQuery, _ := filters[0].(map[string]any)["range"].(map[string]any)
	window, _ := rangeQuery["event.time"].(map[string]any)
	wantStart := time.Unix(100, 0).UTC().Format(time.RFC3339Nano)
	wantEnd := time.Unix(200, 0).UTC().Format(time.RFC3339Nano)
	if window["gte"] != wantStart || window["lte"] != wantEnd {
		t.Fatalf("range = %v", window)
	}
	must, _ := boolQuery["must"].([]any)
	if len(must) != 1 || must[0].(map[string]any)["match"] == nil {
		t.Fatalf("must = %v", must)
	}
}

func TestQueryElasticsearchLogsWrapsPlainQuery(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_search" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		payload, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(payload, &gotBody); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	if _, err := repo.QueryLogs(context.Background(), bo.Datasource{
		UID: "ds-es", Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL,
	}, "level:error", 1, 2, 10, ""); err != nil {
		t.Fatal(err)
	}
	query, _ := gotBody["query"].(map[string]any)
	boolQuery, _ := query["bool"].(map[string]any)
	must, _ := boolQuery["must"].([]any)
	clause, _ := must[0].(map[string]any)
	queryString, _ := clause["query_string"].(map[string]any)
	if queryString["query"] != "level:error" {
		t.Fatalf("query_string = %v", queryString)
	}
	sortFields, _ := gotBody["sort"].([]any)
	sortField, _ := sortFields[0].(map[string]any)
	if order := sortField["@timestamp"].(map[string]any)["order"]; order != "desc" {
		t.Fatalf("sort order = %v", order)
	}
}

func TestQueryElasticsearchLogsRejectsInvalidInput(t *testing.T) {
	repo := newLogTestRepo()
	cases := []bo.Datasource{
		{UID: "ds-es", Driver: "LOGS_ELASTICSEARCH", Endpoint: "http://127.0.0.1:9200", IndexPattern: "logs-*/other"},
		{UID: "ds-es", Driver: "LOGS_ELASTICSEARCH", Endpoint: "http://127.0.0.1:9200", IndexPattern: "logs-*?x=1"},
	}
	for _, item := range cases {
		if _, err := repo.QueryLogs(context.Background(), item, `{"match_all":{}}`, 1, 2, 10, ""); !merr.IsInvalid(err) {
			t.Fatalf("index pattern %q error = %v", item.IndexPattern, err)
		}
	}
	if _, err := esQueryClause(`{"match":`); !merr.IsInvalid(err) {
		t.Fatalf("invalid DSL error = %v", err)
	}
	if _, err := esQueryClause("  "); !merr.IsInvalid(err) {
		t.Fatalf("empty query error = %v", err)
	}
	if _, err := esQueryClause(`{}`); !merr.IsInvalid(err) {
		t.Fatalf("empty DSL error = %v", err)
	}
}

func TestRangeQueryElasticsearchBuildsMatrix(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(payload, &gotBody); err != nil {
			t.Fatalf("request body = %s", payload)
		}
		_, _ = w.Write([]byte(`{"aggregations":{"buckets":{"buckets":[
{"key":1792935600000,"doc_count":3},{"key":1792935615000,"doc_count":0}]}}}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	body, err := repo.RangeQuery(context.Background(), bo.Datasource{
		UID: "ds-es", Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL,
		IndexPattern: "alioth-smoke*", TimeField: "@timestamp",
	}, `{"match":{"level":"error"}}`, 1792935600, 1792935630, 15)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/alioth-smoke*/_search" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["size"] != float64(0) {
		t.Fatalf("size = %v", gotBody["size"])
	}
	var matrix struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Values [][]any `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &matrix); err != nil {
		t.Fatalf("matrix body = %s", body)
	}
	if matrix.Status != "success" || matrix.Data.ResultType != "matrix" || len(matrix.Data.Result) != 1 {
		t.Fatalf("matrix = %s", body)
	}
	values := matrix.Data.Result[0].Values
	if len(values) != 2 || values[0][1] != "3" || values[1][1] != "0" {
		t.Fatalf("values = %v", values)
	}
}

func TestRangeQueryElasticsearchRejectsBadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":{"type":"index_not_found_exception"}}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	if _, err := repo.RangeQuery(context.Background(), bo.Datasource{
		UID: "ds-es", Driver: "LOGS_ELASTICSEARCH", Endpoint: srv.URL, IndexPattern: "missing-*",
	}, "level:error", 0, 30, 15); err == nil {
		t.Fatal("bad response should fail")
	}
}
