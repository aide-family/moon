package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestRangeQueryDoesNotRouteOtherDriversToPrometheus(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
	}))
	defer srv.Close()

	query := NewRangeQuery()
	for _, driver := range []string{"TRACE_JAEGER", "UNKNOWN"} {
		if _, err := query.QueryRange(context.Background(), bo.RuleSnapshot{Driver: driver, Endpoint: srv.URL}, 1, 2, 1); !merr.IsInvalid(err) {
			t.Fatalf("driver %s error = %v, want unsupported-driver error", driver, err)
		}
	}
	if requests != 0 {
		t.Fatalf("unsupported drivers made %d upstream request(s)", requests)
	}
}
