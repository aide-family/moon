package datasource

import (
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestRedactClearsSecrets(t *testing.T) {
	item := redact(&bo.Datasource{Password: "plain-password", Secret: "plain-secret", SecretSet: true})
	if item.Password != "" || item.Secret != "" || !item.SecretSet {
		t.Fatalf("redacted = %+v", item)
	}
}

func TestDriverMatchesLoki(t *testing.T) {
	if !DriverMatches("LOGS", "LOGS_LOKI") {
		t.Fatal("LOGS/LOGS_LOKI should match")
	}
	if DriverMatches("METRICS", "LOGS_LOKI") {
		t.Fatal("METRICS/LOGS_LOKI should not match")
	}
	if DriverMatches("LOGS", "LOGS_UNKNOWN") {
		t.Fatal("unknown logs driver should not match")
	}
}

func TestNormalizeDatasourceKeepsElasticsearchSettings(t *testing.T) {
	item := bo.Datasource{Driver: "LOGS_ELASTICSEARCH", IndexPattern: " logs-* ", TimeField: " event.time "}
	normalizeDatasource(&item)
	if item.IndexPattern != "logs-*" || item.TimeField != "event.time" {
		t.Fatalf("normalized = %+v", item)
	}

	item = bo.Datasource{Driver: "LOGS_LOKI", IndexPattern: "logs-*", TimeField: "event.time"}
	normalizeDatasource(&item)
	if item.IndexPattern != "" || item.TimeField != "" {
		t.Fatalf("loki settings should be cleared: %+v", item)
	}
}

func TestNormalizeLogQuery(t *testing.T) {
	query, size, direction, err := normalizeLogQuery(`{app="api"} |= "error"`, 100, 200, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if query != `{app="api"} |= "error"` || size != defaultLogLimit || direction != "backward" {
		t.Fatalf("normalized = %q %d %q", query, size, direction)
	}

	_, size, direction, err = normalizeLogQuery("{}", 1, 2, maxLogLimit+1, "forward")
	if err != nil {
		t.Fatal(err)
	}
	if size != maxLogLimit || direction != "forward" {
		t.Fatalf("normalized size = %d direction = %q", size, direction)
	}

	cases := []struct {
		name      string
		query     string
		start     int64
		end       int64
		limit     int32
		direction string
	}{
		{name: "empty query", query: "  ", start: 1, end: 2},
		{name: "empty window", query: "{}", start: 2, end: 2},
		{name: "reversed window", query: "{}", start: 3, end: 2},
		{name: "negative start", query: "{}", start: -1, end: 2},
		{name: "unknown direction", query: "{}", start: 1, end: 2, direction: "sideways"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := normalizeLogQuery(tc.query, tc.start, tc.end, tc.limit, tc.direction); !merr.IsInvalid(err) {
				t.Fatalf("normalizeLogQuery error = %v", err)
			}
		})
	}
}
