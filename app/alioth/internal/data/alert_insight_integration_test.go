package data

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// 概览聚合用到 GROUPING SETS、percentile_disc 和跨周表 UNION ALL，
// 只有真库才能验证参数绑定与扫描类型。这组测试只发 SELECT，不会写入任何数据。
//
//	ALIOTH_TEST_DATABASE_URL='postgres://alioth:alioth@127.0.0.1:5432/alioth?sslmode=disable' \
//	  go test ./internal/data -run TestInsight -v
func testDatabaseURL(t *testing.T) string {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("ALIOTH_TEST_DATABASE_URL"))
	if dsn == "" {
		t.Skip("ALIOTH_TEST_DATABASE_URL is not set")
	}
	return dsn
}

func openTestDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping database: %v", err)
	}
	return db
}

func TestInsightAggregateHotTableAgainstRealDatabase(t *testing.T) {
	db := openTestDB(t, testDatabaseURL(t))
	ctx := context.Background()

	var namespaceUID string
	if err := db.QueryRowContext(ctx,
		`SELECT namespace_uid FROM alert_events GROUP BY 1 ORDER BY 1 LIMIT 1`).Scan(&namespaceUID); err != nil {
		t.Skipf("no hot alert events to aggregate: %v", err)
	}
	to := time.Now().Add(24 * time.Hour)
	from := to.AddDate(0, 0, -30)

	query := insightAggregateSQL(insightSource("alert_events", insightColumnStarts, false), insightColumnStarts)
	rows, err := db.QueryContext(ctx, query, "Asia/Shanghai", namespaceUID, from, to)
	if err != nil {
		t.Fatalf("query insight: %v", err)
	}
	items, err := scanInsightRows(rows)
	if err != nil {
		t.Fatalf("scan insight: %v", err)
	}
	if len(items) == 0 {
		t.Skip("no rows in the aggregation window")
	}
	total := 0
	for _, item := range items {
		if item.day == "" {
			total += item.events
		}
		if item.samples > 0 && item.p90 < item.p50 {
			t.Fatalf("p90 %v < p50 %v for %+v", item.p90, item.p50, item)
		}
	}
	if total == 0 {
		t.Fatal("aggregation returned no total row")
	}

	// 闭环序列按 ends_at 分桶，和新增序列是两条独立的时间轴。
	closedQuery := insightAggregateSQL(insightSource("alert_events", insightColumnEnds, false), insightColumnEnds)
	closedRows, err := db.QueryContext(ctx, closedQuery, "Asia/Shanghai", namespaceUID, from, to)
	if err != nil {
		t.Fatalf("query closed insight: %v", err)
	}
	closedItems, err := scanInsightRows(closedRows)
	if err != nil {
		t.Fatalf("scan closed insight: %v", err)
	}
	for _, item := range closedItems {
		if item.day != "" && item.status != "" && item.samples != item.events {
			t.Fatalf("closed row should always have an end time: %+v", item)
		}
	}
}

func TestHistoryInsightWithExclusionAgainstRealDatabase(t *testing.T) {
	db := openTestDB(t, testDatabaseURL(t))
	ctx := context.Background()

	history, table, namespaceUID := firstHistoryShard(t, db)
	from := time.Now().AddDate(0, 0, -30)
	to := time.Now().Add(24 * time.Hour)
	source := insightSource(quoteShardIdent(table), insightColumnEnds, true)
	rows, err := history.QueryContext(ctx, insightAggregateSQL(source, insightColumnEnds), "Asia/Shanghai",
		namespaceUID, from, to, []string{"does-not-exist"})
	if err != nil {
		t.Fatalf("query history insight: %v", err)
	}
	items, err := scanInsightRows(rows)
	if err != nil {
		t.Fatalf("scan history insight: %v", err)
	}
	for _, item := range items {
		if item.day == "" && item.events > 0 {
			if item.samples > 0 && item.p90 < item.p50 {
				t.Fatalf("p90 %v < p50 %v for %+v", item.p90, item.p50, item)
			}
			return
		}
	}
	t.Skip("no rows in the history window")
}

// firstHistoryShard 找到第一个真正有数据的租户周表。开发库里可能有空的周表和空租户库。
func firstHistoryShard(t *testing.T, db *sql.DB) (*sql.DB, string, string) {
	t.Helper()
	ctx := context.Background()
	rows, err := db.QueryContext(ctx,
		`SELECT datname FROM pg_database WHERE datname LIKE 'alioth_alert_%' ORDER BY datname`)
	if err != nil {
		t.Fatalf("list tenant databases: %v", err)
	}
	databases := make([]string, 0, 4)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			t.Fatalf("scan tenant database: %v", err)
		}
		databases = append(databases, name)
	}
	_ = rows.Close()
	for _, database := range databases {
		history := openTestDB(t, withDatabase(t, os.Getenv("ALIOTH_TEST_DATABASE_URL"), database))
		tableRows, err := history.QueryContext(ctx,
			`SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'alert_events_%'
			 ORDER BY tablename DESC`)
		if err != nil {
			continue
		}
		tables := make([]string, 0, 8)
		for tableRows.Next() {
			var name string
			if err := tableRows.Scan(&name); err != nil {
				_ = tableRows.Close()
				break
			}
			tables = append(tables, name)
		}
		_ = tableRows.Close()
		for _, table := range tables {
			var namespaceUID string
			if err := history.QueryRowContext(ctx,
				`SELECT namespace_uid FROM `+quoteShardIdent(table)+` GROUP BY 1 ORDER BY 1 LIMIT 1`).Scan(&namespaceUID); err != nil {
				continue
			}
			return history, table, namespaceUID
		}
	}
	t.Skip("no tenant weekly table has rows")
	return nil, "", ""
}

// withDatabase 把 DSN 的库名换成另一个租户库，其余连接参数保持不变。
func withDatabase(t *testing.T, dsn, database string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Host == "" {
		t.Fatalf("ALIOTH_TEST_DATABASE_URL must be a URL DSN, got %q", dsn)
	}
	parsed.Path = "/" + database
	return parsed.String()
}
