package data

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestHistoryShardNames(t *testing.T) {
	store := &historyStore{shardStore: &shardStore{dbPrefix: "alioth_alert_"}, tablePrefix: "alert_events_"}
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
	if !validShardPrefix(strings.Repeat("a", maxShardDatabasePrefixLen), maxShardDatabasePrefixLen) {
		t.Fatal("maximum database prefix length should be valid")
	}
	if validShardPrefix(strings.Repeat("a", maxShardDatabasePrefixLen+1), maxShardDatabasePrefixLen) {
		t.Fatal("database prefix longer than the maximum should be invalid")
	}
	if !validShardPrefix(strings.Repeat("a", maxShardTablePrefixLen), maxShardTablePrefixLen) {
		t.Fatal("maximum table prefix length should be valid")
	}
	if validShardPrefix(strings.Repeat("a", maxShardTablePrefixLen+1), maxShardTablePrefixLen) {
		t.Fatal("table prefix longer than the maximum should be invalid")
	}

	store := newHistoryStore(
		newShardStore(nil, "", strings.Repeat("a", maxShardDatabasePrefixLen+1), nil),
		strings.Repeat("b", maxShardTablePrefixLen+1),
	)
	if store.dbPrefix != defaultShardDatabasePrefix || store.tablePrefix != defaultHistoryTablePrefix {
		t.Fatalf("invalid prefixes should fall back to defaults: db=%q table=%q", store.dbPrefix, store.tablePrefix)
	}

	namespaceUID := "1b8307a4abf29e8f783bb9c0dd651695"
	if got := len(shardDatabaseName(strings.Repeat("a", maxShardDatabasePrefixLen), namespaceUID)); got != shardIdentMaxLen {
		t.Fatalf("maximum database prefix should produce a %d-character name, got %d", shardIdentMaxLen, got)
	}
	store = &historyStore{shardStore: &shardStore{}, tablePrefix: strings.Repeat("a", maxShardTablePrefixLen)}
	tableName, _ := store.tableName(time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC))
	if got := len(tableName); got != shardIdentMaxLen {
		t.Fatalf("maximum table prefix should produce a %d-character name, got %d", shardIdentMaxLen, got)
	}
}

func TestHistoryDSN(t *testing.T) {
	got, err := shardDSN("postgres://alioth:alioth@127.0.0.1:5432/alioth?sslmode=disable", "alioth_alert_ns")
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
