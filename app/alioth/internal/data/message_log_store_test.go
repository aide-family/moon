package data

import (
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestMessageLogShardNames(t *testing.T) {
	store := newMessageLogStore(&shardStore{dbPrefix: "alioth_alert_"}, "message_logs_")
	if got := store.databaseName("1b8307a4abf29e8f783bb9c0dd651695"); got != "alioth_alert_1b8307a4abf29e8f783bb9c0dd651695" {
		t.Fatalf("database name = %s", got)
	}
	// 周四归一到周一，与告警历史的分片规则一致。
	thursday := time.Date(2026, 10, 22, 13, 0, 0, 0, time.UTC)
	table, week := store.tableName(thursday)
	if table != "message_logs_20261019" {
		t.Fatalf("table name = %s", table)
	}
	if week.Weekday() != time.Monday || week.Format("20060102") != "20261019" {
		t.Fatalf("week = %s", week)
	}
}

func TestMessageLogPrefixFallback(t *testing.T) {
	store := newMessageLogStore(
		newShardStore(nil, "", "", nil),
		strings.Repeat("a", maxShardTablePrefixLen+1),
	)
	if store.dbPrefix != defaultShardDatabasePrefix || store.tablePrefix != defaultMessageLogTablePrefix {
		t.Fatalf("invalid prefixes should fall back to defaults: db=%q table=%q", store.dbPrefix, store.tablePrefix)
	}
}

func TestFilterMessageLogTables(t *testing.T) {
	names := []string{
		"alert_events_20261019",
		"message_logs_20261012",
		"message_logs_20261019",
		"message_logs_20261026",
		"message_logs_bad",
	}
	start := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	got := filterMessageLogTables(names, "message_logs_", start, end)
	want := []string{"message_logs_20261012", "message_logs_20261019"}
	if len(got) != len(want) {
		t.Fatalf("tables = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tables = %v", got)
		}
	}
	// 终点正好是周一零点时，新一周的表不参与查询（EndAt 不包含）。
	got = filterMessageLogTables(names, "message_logs_", time.Time{}, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	if len(got) != 1 || got[0] != "message_logs_20261012" {
		t.Fatalf("tables before monday = %v", got)
	}
}

func TestMergeMessageLogs(t *testing.T) {
	base := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	sharded := []bo.MessageLog{
		{UID: "a", CreatedAt: base},
		{UID: "c", CreatedAt: base.Add(time.Hour)},
	}
	fallback := []bo.MessageLog{
		{UID: "a", CreatedAt: base},
		{UID: "b", CreatedAt: base.Add(30 * time.Minute)},
	}
	got := mergeMessageLogs(sharded, fallback, 10)
	if len(got) != 3 {
		t.Fatalf("merged = %d", len(got))
	}
	order := []string{got[0].UID, got[1].UID, got[2].UID}
	want := []string{"c", "b", "a"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v", order)
		}
	}
	if limited := mergeMessageLogs(sharded, fallback, 2); len(limited) != 2 || limited[0].UID != "c" {
		t.Fatalf("limited = %v", limited)
	}
}

func TestMessageLogShardWhere(t *testing.T) {
	start := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	where, args := messageLogShardWhere("ns", bo.MessageLogListFilter{
		TaskUID: "task", Status: "FAILED", StartAt: start, EndAt: end,
	})
	if where != "namespace_uid = $1 AND task_uid = $2 AND status = $3 AND created_at >= $4 AND created_at < $5" {
		t.Fatalf("where = %s", where)
	}
	if len(args) != 5 || args[0] != "ns" || args[1] != "task" || args[2] != "FAILED" {
		t.Fatalf("args = %v", args)
	}
}
