package esagg

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSearchPath(t *testing.T) {
	cases := []struct {
		pattern string
		want    string
	}{
		{"", "/_search"},
		{"  ", "/_search"},
		{"alioth-smoke*", "/alioth-smoke*/_search"},
		{"app-2026.10.*,.internal-*", "/app-2026.10.*,.internal-*/_search"},
	}
	for _, item := range cases {
		got, err := SearchPath(item.pattern)
		if err != nil {
			t.Fatalf("SearchPath(%q) error = %v", item.pattern, err)
		}
		if got != item.want {
			t.Fatalf("SearchPath(%q) = %q, want %q", item.pattern, got, item.want)
		}
	}
	for _, pattern := range []string{"../etc", "index/other", "a b", strings.Repeat("a", 257)} {
		if _, err := SearchPath(pattern); err == nil {
			t.Fatalf("SearchPath(%q) should fail", pattern)
		}
	}
}

func TestBodyCountsDocumentsPerBucket(t *testing.T) {
	payload, err := Body(Query{
		Expr:         `{"match":{"level":"error"}}`,
		IndexPattern: "alioth-smoke*",
		TimeField:    "@timestamp",
		Start:        1792937600,
		End:          1792937630,
		Step:         15,
	})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body["size"] != float64(0) {
		t.Fatalf("size = %v, want 0（只要聚合结果）", body["size"])
	}
	query := body["query"].(map[string]any)["bool"].(map[string]any)
	filter := query["filter"].([]any)[0].(map[string]any)["range"].(map[string]any)["@timestamp"].(map[string]any)
	if filter["gte"] != float64(1792937600000) || filter["lt"] != float64(1792937630000) {
		t.Fatalf("range = %v，应为左闭右开", filter)
	}
	must := query["must"].([]any)[0].(map[string]any)
	if _, ok := must["match"]; !ok {
		t.Fatalf("must = %v，DSL 应原样透传", must)
	}
	histogram := body["aggs"].(map[string]any)[aggName].(map[string]any)["date_histogram"].(map[string]any)
	if histogram["field"] != "@timestamp" || histogram["fixed_interval"] != "15s" {
		t.Fatalf("date_histogram = %v", histogram)
	}
	if histogram["min_doc_count"] != float64(0) {
		t.Fatalf("min_doc_count = %v，空桶也要返回", histogram["min_doc_count"])
	}
	bounds := histogram["extended_bounds"].(map[string]any)
	if bounds["min"] != float64(1792937600000) || bounds["max"] != float64(1792937629999) {
		t.Fatalf("extended_bounds = %v", bounds)
	}
}

func TestBodyWrapsPlainQuery(t *testing.T) {
	payload, err := Body(Query{Expr: `level:error and app:"api"`, Start: 100, End: 130, Step: 15})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	must := body["query"].(map[string]any)["bool"].(map[string]any)["must"].([]any)[0].(map[string]any)
	queryString, ok := must["query_string"].(map[string]any)
	if !ok {
		t.Fatalf("must = %v", must)
	}
	if queryString["query"] != `level:error and app:"api"` {
		t.Fatalf("query_string = %v", queryString)
	}
	// 空时间字段回落到默认值。
	histogram := body["aggs"].(map[string]any)[aggName].(map[string]any)["date_histogram"].(map[string]any)
	if histogram["field"] != DefaultTimeField {
		t.Fatalf("field = %v", histogram["field"])
	}
}

func TestBodyRejectsBadInput(t *testing.T) {
	cases := []Query{
		{Expr: "", Start: 0, End: 15, Step: 15},
		{Expr: "{}", Start: 0, End: 15, Step: 15},
		{Expr: "{not json}", Start: 0, End: 15, Step: 15},
		{Expr: "ok", Start: 15, End: 15, Step: 15},
		{Expr: "ok", Start: 0, End: 15, TimeField: "bad field"},
	}
	for _, item := range cases {
		if _, err := Body(item); err == nil {
			t.Fatalf("Body(%+v) should fail", item)
		}
	}
}

func TestParseBuckets(t *testing.T) {
	payload := []byte(`{"aggregations":{"buckets":{"buckets":[
{"key_as_string":"2026-10-25T14:00:00.000Z","key":1792935600000,"doc_count":0},
{"key_as_string":"2026-10-25T14:00:15.000Z","key":1792935615000,"doc_count":7}]}}}`)
	buckets, err := ParseBuckets(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets) != 2 {
		t.Fatalf("buckets = %+v", buckets)
	}
	if !buckets[0].At.Equal(time.UnixMilli(1792935600000).UTC()) || buckets[0].Count != 0 {
		t.Fatalf("bucket[0] = %+v", buckets[0])
	}
	if buckets[1].Count != 7 || buckets[1].Raw != "7" {
		t.Fatalf("bucket[1] = %+v", buckets[1])
	}
}

func TestParseBucketsRejectsBadPayload(t *testing.T) {
	for _, payload := range []string{
		``,
		`{"error":{"type":"index_not_found_exception"}}`,
		`{"aggregations":{}}`,
		`{"aggregations":{"buckets":{"buckets":[{"doc_count":1}]}}}`,
	} {
		if _, err := ParseBuckets([]byte(payload)); err == nil {
			t.Fatalf("ParseBuckets(%q) should fail：查询失败不能被当成零命中", payload)
		}
	}
}

func TestMatrixShape(t *testing.T) {
	payload, err := Matrix([]Bucket{
		{At: time.Unix(1792937600, 0).UTC(), Count: 3},
		{At: time.Unix(1792937615, 0).UTC(), Count: 0},
	}, map[string]string{"index": "alioth-smoke"})
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]any           `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "success" || body.Data.ResultType != "matrix" || len(body.Data.Result) != 1 {
		t.Fatalf("body = %+v", body)
	}
	series := body.Data.Result[0]
	if series.Metric["index"] != "alioth-smoke" || len(series.Values) != 2 {
		t.Fatalf("series = %+v", series)
	}
	if series.Values[0][1] != "3" || series.Values[1][1] != "0" {
		t.Fatalf("values = %v", series.Values)
	}
}
