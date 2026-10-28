package datasource

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
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
	if !DriverMatches("LOGS", "LOGS_VLOGS") {
		t.Fatal("LOGS/LOGS_VLOGS should match")
	}
	if DriverMatches("METRICS", "LOGS_LOKI") {
		t.Fatal("METRICS/LOGS_LOKI should not match")
	}
	if DriverMatches("LOGS", "LOGS_UNKNOWN") {
		t.Fatal("unknown logs driver should not match")
	}
}

func TestValidateDatasourceEndpoint(t *testing.T) {
	valid := []string{
		"http://127.0.0.1:9090",
		"https://metrics.example.com/prometheus",
		" http://victoriametrics.moon.com/select/0/prometheus ",
	}
	for _, endpoint := range valid {
		if err := validateDatasource(bo.Datasource{Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: endpoint}); err != nil {
			t.Fatalf("validateDatasource(%q) error = %v", endpoint, err)
		}
	}

	invalid := []string{
		"",
		"htp://es.moon.com",
		"es.moon.com",
		"ftp://es.moon.com",
		"http://",
		"http://exa mple.com",
	}
	for _, endpoint := range invalid {
		if err := validateDatasource(bo.Datasource{Name: "es", Type: "LOGS", Driver: "LOGS_ELASTICSEARCH", Endpoint: endpoint}); !merr.IsInvalid(err) {
			t.Fatalf("validateDatasource(%q) error = %v", endpoint, err)
		}
	}

	if err := validateDatasource(bo.Datasource{Name: "bad", Type: "METRICS", Driver: "LOGS_LOKI", Endpoint: "http://loki"}); !merr.IsInvalid(err) {
		t.Fatalf("mismatched driver error = %v", err)
	}
	if err := validateDatasource(bo.Datasource{Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: "http://prom"}); !merr.IsInvalid(err) {
		t.Fatalf("empty name error = %v", err)
	}
	if err := validateDatasource(bo.Datasource{Name: strings.Repeat("长", maxDatasourceName+1), Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: "http://prom"}); !merr.IsInvalid(err) {
		t.Fatalf("long name error = %v", err)
	}
}

func TestNormalizeDatasourceEndpoint(t *testing.T) {
	item := bo.Datasource{Name: " prom ", Endpoint: " HTTP://Prometheus.Moon.com:9090/ "}
	normalizeDatasource(&item)
	if item.Name != "prom" || item.Endpoint != "http://prometheus.moon.com:9090" {
		t.Fatalf("normalized = %+v", item)
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

// availabilityDatasource 生成一条地址和驱动都合法、只用于测试可用性字段的数据源。
func availabilityDatasource() bo.Datasource {
	return bo.Datasource{
		Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: "http://127.0.0.1:9090",
	}
}

func TestNormalizeDatasourceAvailabilityDefaults(t *testing.T) {
	item := availabilityDatasource()
	normalizeDatasource(&item)
	if item.FailureWindowMinutes != defaultFailureWindowMinutes || item.FailureThreshold != defaultFailureThreshold {
		t.Fatalf("defaults = %d/%d", item.FailureWindowMinutes, item.FailureThreshold)
	}
	if item.NotificationGroupUIDs == nil || len(item.NotificationGroupUIDs) != 0 {
		t.Fatalf("group uids = %#v, want empty slice", item.NotificationGroupUIDs)
	}
	if err := validateDatasource(item); err != nil {
		t.Fatalf("disabled defaults should validate: %v", err)
	}
}

func TestNormalizeDatasourceNotificationGroups(t *testing.T) {
	item := availabilityDatasource()
	item.NotificationGroupUIDs = []string{" group-1 ", "", "group-2", "group-1"}
	normalizeDatasource(&item)
	if len(item.NotificationGroupUIDs) != 2 || item.NotificationGroupUIDs[0] != "group-1" || item.NotificationGroupUIDs[1] != "group-2" {
		t.Fatalf("group uids = %#v", item.NotificationGroupUIDs)
	}
}

func TestValidateDatasourceAvailability(t *testing.T) {
	valid := []bo.Datasource{
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 1, FailureThreshold: 1},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: maxFailureWindowMinutes, FailureThreshold: maxFailureThreshold},
		{AvailabilityAlertEnabled: false, FailureWindowMinutes: 0, FailureThreshold: 0},
	}
	for _, tc := range valid {
		item := availabilityDatasource()
		item.AvailabilityAlertEnabled = tc.AvailabilityAlertEnabled
		item.FailureWindowMinutes = tc.FailureWindowMinutes
		item.FailureThreshold = tc.FailureThreshold
		if err := validateDatasource(item); err != nil {
			t.Fatalf("validate(%+v) error = %v", tc, err)
		}
	}

	invalid := []bo.Datasource{
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 0, FailureThreshold: 3},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 5, FailureThreshold: 0},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: maxFailureWindowMinutes + 1, FailureThreshold: 3},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 5, FailureThreshold: maxFailureThreshold + 1},
		{AvailabilityAlertEnabled: false, FailureWindowMinutes: -1, FailureThreshold: 3},
	}
	for _, tc := range invalid {
		item := availabilityDatasource()
		item.AvailabilityAlertEnabled = tc.AvailabilityAlertEnabled
		item.FailureWindowMinutes = tc.FailureWindowMinutes
		item.FailureThreshold = tc.FailureThreshold
		if err := validateDatasource(item); !merr.IsInvalid(err) {
			t.Fatalf("validate(%+v) error = %v", tc, err)
		}
	}

	tooMany := availabilityDatasource()
	tooMany.NotificationGroupUIDs = make([]string, maxNotificationGroups+1)
	if err := validateDatasource(tooMany); !merr.IsInvalid(err) {
		t.Fatalf("too many groups error = %v", err)
	}
}
