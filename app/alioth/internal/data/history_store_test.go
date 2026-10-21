package data

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHistoryShardNames(t *testing.T) {
	store := &historyStore{dbPrefix: "alioth_alert_", tablePrefix: "alert_events_"}
	if got := store.databaseName("1b8307a4abf29e8f783bb9c0dd651695"); got != "alioth_alert_1b8307a4abf29e8f783bb9c0dd651695" {
		t.Fatalf("database name = %s", got)
	}
	// 周四归一到周一。
	thursday := time.Date(2026, 10, 22, 13, 0, 0, 0, time.UTC)
	table, week := store.tableName(thursday)
	if table != "alert_events_20261019" {
		t.Fatalf("table name = %s", table)
	}
	if week.Weekday() != time.Monday || week.Format("20060102") != "20261019" {
		t.Fatalf("week = %s", week)
	}
}

func TestHistoryDSN(t *testing.T) {
	got, err := historyDSN("postgres://alioth:alioth@127.0.0.1:5432/alioth?sslmode=disable", "alioth_alert_ns")
	if err != nil {
		t.Fatal(err)
	}
	if got != "postgres://alioth:alioth@127.0.0.1:5432/alioth_alert_ns?sslmode=disable" {
		t.Fatalf("dsn = %s", got)
	}
}

func TestMergeEventLabelsSamplePriority(t *testing.T) {
	source := datasourceMetaInfo{Name: "prom", Labels: `{"env":"prod"}`, LevelUID: "ds-level"}
	levels := map[string]string{"alert-level": "P1", "ds-level": "P2"}
	merged := mergeEventLabels(`{"datasource":"from-sample","env":"staging"}`, source, "alert-level", levels)
	got := map[string]string{}
	if err := json.Unmarshal([]byte(merged), &got); err != nil {
		t.Fatal(err)
	}
	if got["datasource"] != "from-sample" || got["env"] != "staging" {
		t.Fatalf("sample labels lost: %s", merged)
	}
	if got["alert_level"] != "P1" || got["datasource_level"] != "P2" {
		t.Fatalf("metadata labels missing: %s", merged)
	}
}
