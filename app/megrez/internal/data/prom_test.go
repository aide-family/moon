package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestQueryRangeKeepsEndpointPrefix(t *testing.T) {
	var (
		gotPath, gotQuery, gotStart, gotEnd, gotStep string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("query")
		gotStart = r.URL.Query().Get("start")
		gotEnd = r.URL.Query().Get("end")
		gotStep = r.URL.Query().Get("step")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up"},"values":[[100,"1"],[115,"2"]]}]}}`))
	}))
	defer srv.Close()

	samples, err := NewRangeQuery().QueryRange(context.Background(), bo.RuleSnapshot{
		UID: "rule-1", Endpoint: srv.URL + "/select/0/prometheus", Expr: "up",
	}, 100, 200, 15)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/select/0/prometheus/api/v1/query_range" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "up" || gotStart != "100" || gotEnd != "200" || gotStep != "15" {
		t.Fatalf("query = %q start/end/step = %q/%q/%q", gotQuery, gotStart, gotEnd, gotStep)
	}
	if len(samples) != 1 || samples[0].Labels["__name__"] != "up" || len(samples[0].Points) != 2 {
		t.Fatalf("samples = %+v", samples)
	}
	if samples[0].Points[1].Value != 2 || samples[0].Points[1].Raw != "2" {
		t.Fatalf("point = %+v", samples[0].Points[1])
	}
}

func TestQueryRangeRejectsInvalidEndpoint(t *testing.T) {
	for _, endpoint := range []string{"", "127.0.0.1:9090", "ftp://127.0.0.1"} {
		if _, err := NewRangeQuery().QueryRange(context.Background(), bo.RuleSnapshot{
			UID: "rule-1", Endpoint: endpoint, Expr: "up",
		}, 100, 200, 15); !merr.IsInvalid(err) {
			t.Fatalf("endpoint %q error = %v", endpoint, err)
		}
	}
}

func TestDecodeMatrixRejectsInvalidPayload(t *testing.T) {
	if _, err := decodeMatrix([]byte(`{"status":"error"}`)); err == nil {
		t.Fatal("error payload should be rejected")
	}
	if _, err := decodeMatrix([]byte(`not-json`)); err == nil {
		t.Fatal("non-json payload should be rejected")
	}
	if _, err := decodeMatrix([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`)); err == nil {
		t.Fatal("vector result should be rejected")
	}
}

func TestDecodeMatrixSkipsInvalidPoints(t *testing.T) {
	samples, err := decodeMatrix([]byte(`{"status":"success","data":{"resultType":"matrix","result":[
		{"metric":{"__name__":"up"},"values":[[100,"1"],[115,2],[130,"nan"]]}
	]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 || len(samples[0].Points) != 2 {
		t.Fatalf("samples = %+v", samples)
	}
	if samples[0].Points[0].At.Unix() != 100 || samples[0].Points[1].Raw != "nan" {
		t.Fatalf("points = %+v", samples[0].Points)
	}
}
