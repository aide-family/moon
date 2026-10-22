package data

import (
	"encoding/json"
	"strings"
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

func TestHistoryPrefixLengthLimits(t *testing.T) {
	if !validHistoryPrefix(strings.Repeat("a", maxHistoryDatabasePrefixLen), maxHistoryDatabasePrefixLen) {
		t.Fatal("maximum database prefix length should be valid")
	}
	if validHistoryPrefix(strings.Repeat("a", maxHistoryDatabasePrefixLen+1), maxHistoryDatabasePrefixLen) {
		t.Fatal("database prefix longer than the maximum should be invalid")
	}
	if !validHistoryPrefix(strings.Repeat("a", maxHistoryTablePrefixLen), maxHistoryTablePrefixLen) {
		t.Fatal("maximum table prefix length should be valid")
	}
	if validHistoryPrefix(strings.Repeat("a", maxHistoryTablePrefixLen+1), maxHistoryTablePrefixLen) {
		t.Fatal("table prefix longer than the maximum should be invalid")
	}

	store := newHistoryStore(nil, "", strings.Repeat("a", maxHistoryDatabasePrefixLen+1), strings.Repeat("b", maxHistoryTablePrefixLen+1), nil)
	if store.dbPrefix != defaultHistoryDatabasePrefix || store.tablePrefix != defaultHistoryTablePrefix {
		t.Fatalf("invalid prefixes should fall back to defaults: db=%q table=%q", store.dbPrefix, store.tablePrefix)
	}

	namespaceUID := "1b8307a4abf29e8f783bb9c0dd651695"
	if got := len(historyDatabaseName(strings.Repeat("a", maxHistoryDatabasePrefixLen), namespaceUID)); got != historyIdentMaxLen {
		t.Fatalf("maximum database prefix should produce a %d-character name, got %d", historyIdentMaxLen, got)
	}
	store = &historyStore{tablePrefix: strings.Repeat("a", maxHistoryTablePrefixLen)}
	tableName, _ := store.tableName(time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC))
	if got := len(tableName); got != historyIdentMaxLen {
		t.Fatalf("maximum table prefix should produce a %d-character name, got %d", historyIdentMaxLen, got)
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
