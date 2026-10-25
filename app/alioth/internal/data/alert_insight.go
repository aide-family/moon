package data

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

// 每种事件只能是「进」或「出」的一端：
// 新增按 starts_at 分桶，闭环按 ends_at 分桶。两条时间轴必须分开算，
// 否则区间内的「恢复数」会变成「当日新增里最终恢复了多少」，闭环与新增对不上账。
const (
	insightColumnStarts = "starts_at"
	insightColumnEnds   = "ends_at"
)

const (
	// insightTableCushion 是新增序列的周表裁剪余量。
	// 周表按 created_at（与 starts_at 同刻）的 UTC 周命名，极少数跨周边界的事件
	// 会落在相邻周表里，多扫一周的成本可以忽略。
	insightTableCushion = 7 * 24 * time.Hour
	// insightClosedLookback 是闭环序列的回溯窗口。
	// 归档周表按「创建周」分表，一条三周前触发、今天才闭环的事件仍在三周前那张表里；
	// 没有 ends_at 索引时只能按创建周回溯，30 天覆盖绝大多数事件生命周期。
	insightClosedLookback = 30 * 24 * time.Hour
	// insightHotUIDLimit 是聚合时从主库取回用于去重的 uid 上限。
	// 主库只保留仍在触发（以及归档失败的兜底行），正常远小于这个值。
	insightHotUIDLimit = 5000
)

// insightRow 是区间聚合的一行。
// day 为空表示区间汇总行；status 为空表示按天汇总行（只提供该日的分位数）。
type insightRow struct {
	day      string
	status   string
	levelUID string
	events   int
	samples  int
	p50      float64
	p90      float64
}

// insightAggregateSQL 在给定事件子集上按本地自然日、状态、级别聚合。
//
// 口径说明：
//   - 分桶列按调用方选择（starts_at / ends_at），换成本地日期；
//   - 中位数用 percentile_cont(0.5)，偶数样本取两中值平均；
//   - P90 必须用 percentile_disc(0.9)，对应最近秩 sorted[ceil(0.9n)-1]，
//     与历史分析页的 durationSummary 完全一致，不能换成 percentile_cont；
//   - GROUPING SETS 一次查询同时给出（天×状态×级别）、（天）和（区间）三种粒度。
func insightAggregateSQL(source, column string) string {
	if column != insightColumnStarts && column != insightColumnEnds {
		// 列名只由本包的常量传入，这里兜底防止拼接出错。
		column = insightColumnStarts
	}
	bucket := fmt.Sprintf("date_trunc('day', %s AT TIME ZONE $1)", column)
	return fmt.Sprintf(`SELECT
		coalesce(to_char(%s, 'YYYY-MM-DD'), '') AS day,
		coalesce(status, '') AS status,
		coalesce(level_uid, '') AS level_uid,
		count(*) AS events,
		count(*) FILTER (WHERE ends_at IS NOT NULL) AS samples,
		coalesce(percentile_cont(0.5) WITHIN GROUP (ORDER BY extract(epoch FROM ends_at - starts_at))
			FILTER (WHERE ends_at IS NOT NULL), 0) AS p50,
		coalesce(percentile_disc(0.9) WITHIN GROUP (ORDER BY extract(epoch FROM ends_at - starts_at))
			FILTER (WHERE ends_at IS NOT NULL), 0) AS p90
	FROM (%s) e
	GROUP BY GROUPING SETS (
		(%s, status, level_uid),
		(%s),
		()
	)`, bucket, source, bucket, bucket)
}

// insightSource 构造聚合的源集合：只取 column 落在 [$3, $4) 的行。
// exclude 为真时把主库里已存在的事件 uid 排除，避免实时表与归档表重复计数。
func insightSource(relation, column string, exclude bool) string {
	if column != insightColumnStarts && column != insightColumnEnds {
		column = insightColumnStarts
	}
	filter := ""
	if exclude {
		filter = " AND NOT (uid = ANY($5))"
	}
	return fmt.Sprintf(`SELECT starts_at, ends_at, status, level_uid FROM %s
		WHERE namespace_uid = $2 AND %s >= $3 AND %s < $4%s`, relation, column, column, filter)
}

func scanInsightRows(rows *sql.Rows) ([]insightRow, error) {
	defer func() { _ = rows.Close() }()
	items := make([]insightRow, 0, 64)
	for rows.Next() {
		var (
			row insightRow
			p50 sql.NullFloat64
			p90 sql.NullFloat64
		)
		if err := rows.Scan(&row.day, &row.status, &row.levelUID, &row.events, &row.samples, &p50, &p90); err != nil {
			return nil, err
		}
		row.p50 = p50.Float64
		row.p90 = p90.Float64
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// Insight 聚合区间内的告警：实时表与归档周表分别聚合后合并。
// 两个库不能 JOIN，所以主库在窗口内的 uid 会被带给归档查询用于去重，主库行优先。
func (r *alertEventRepo) Insight(ctx context.Context, namespaceUID string, from, to time.Time, loc *time.Location) (bo.AlertInsight, error) {
	if loc == nil {
		loc = time.UTC
	}
	startedHot, err := r.insightHot(ctx, namespaceUID, from, to, loc, insightColumnStarts)
	if err != nil {
		return bo.AlertInsight{}, err
	}
	closedHot, err := r.insightHot(ctx, namespaceUID, from, to, loc, insightColumnEnds)
	if err != nil {
		return bo.AlertInsight{}, err
	}
	startedArchived := []insightRow{}
	closedArchived := []insightRow{}
	complete := r.history == nil
	if r.history != nil {
		uids, err := r.insightHotUIDs(ctx, namespaceUID, from, to)
		if err != nil {
			return bo.AlertInsight{}, err
		}
		startedArchived, err = r.history.insight(ctx, namespaceUID,
			from.Add(-insightTableCushion), from, to, loc, uids, insightColumnStarts)
		if err == nil {
			closedArchived, err = r.history.insight(ctx, namespaceUID,
				from.Add(-insightClosedLookback), from, to, loc, uids, insightColumnEnds)
		}
		if err != nil {
			// 分库不可用时退回实时表结果，概览页降级而不是整页报错。
			r.log.Error("aggregate alert history: " + err.Error())
			startedArchived = []insightRow{}
			closedArchived = []insightRow{}
		} else {
			complete = true
		}
	}
	return buildInsight(startedHot, startedArchived, closedHot, closedArchived, from, to, loc, complete), nil
}

func (r *alertEventRepo) insightHot(ctx context.Context, namespaceUID string, from, to time.Time, loc *time.Location, column string) ([]insightRow, error) {
	query := insightAggregateSQL(insightSource("alert_events", column, false), column)
	rows, err := r.db.QueryContext(ctx, query, loc.String(), namespaceUID, from, to)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := scanInsightRows(rows)
	if err != nil {
		return nil, r.fail(err)
	}
	return items, nil
}

// insightHotUIDs 取主库窗口内的事件标识：两条时间轴都要覆盖，
// 因为新增序列按 starts_at、闭环序列按 ends_at 与归档去重。
func (r *alertEventRepo) insightHotUIDs(ctx context.Context, namespaceUID string, from, to time.Time) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT uid FROM alert_events
		WHERE namespace_uid = $1
		  AND ((starts_at >= $2 AND starts_at < $3) OR (ends_at >= $2 AND ends_at < $3))
		LIMIT $4`, namespaceUID, from, to, insightHotUIDLimit)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = rows.Close() }()
	uids := make([]string, 0, 64)
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, r.fail(err)
		}
		uids = append(uids, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, r.fail(err)
	}
	return uids, nil
}

// insight 在租户库的周表上做同样的聚合。
// tableFrom 决定扫哪些周表（周表按创建周命名，闭环序列要往前多看一段），
// from/to 决定取哪些行（按 column 对应的时刻裁剪）。
func (s *historyStore) insight(ctx context.Context, namespaceUID string, tableFrom, from, to time.Time, loc *time.Location, exclude []string, column string) ([]insightRow, error) {
	exists, err := s.tenantExists(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []insightRow{}, nil
	}
	conn, _, err := s.ensureTenant(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	rows, err := conn.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE $1 ORDER BY tablename`, s.tablePrefix+"%")
	if err != nil {
		s.log.Error(fmt.Sprintf("history tables %s: %v", s.databaseName(namespaceUID), err))
		return nil, merr.Server("history storage is unavailable")
	}
	tables := make([]string, 0, 8)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			_ = rows.Close()
			return nil, merr.Server("history storage is unavailable")
		}
		if !shardIdentPattern.MatchString(table) {
			continue
		}
		if !historyTableOverlaps(table, s.tablePrefix, tableFrom, to) {
			continue
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, merr.Server("history storage is unavailable")
	}
	_ = rows.Close()
	if len(tables) == 0 {
		return []insightRow{}, nil
	}

	args := []any{loc.String(), namespaceUID, from, to}
	withExclude := len(exclude) > 0
	if withExclude {
		args = append(args, exclude)
	}
	branches := make([]string, 0, len(tables))
	for _, table := range tables {
		branches = append(branches, insightSource(quoteShardIdent(table), column, withExclude))
	}
	result, err := conn.QueryContext(ctx, insightAggregateSQL(strings.Join(branches, "\nUNION ALL\n"), column), args...)
	if err != nil {
		s.log.Error(fmt.Sprintf("history insight %s: %v", s.databaseName(namespaceUID), err))
		return nil, merr.Server("history storage is unavailable")
	}
	items, err := scanInsightRows(result)
	if err != nil {
		s.log.Error(fmt.Sprintf("history insight scan %s: %v", s.databaseName(namespaceUID), err))
		return nil, merr.Server("history storage is unavailable")
	}
	return items, nil
}

// historyTableOverlaps 判断某张周表是否覆盖 [from, to) 中的任意时刻。
// 周表名后缀是该周周一的 UTC 日期，一张表覆盖 7 天。
func historyTableOverlaps(table, prefix string, from, to time.Time) bool {
	suffix := strings.TrimPrefix(table, prefix)
	weekStart, err := time.ParseInLocation("20060102", suffix, time.UTC)
	if err != nil {
		return false
	}
	weekEnd := weekStart.AddDate(0, 0, 7)
	return weekEnd.After(from.UTC()) && weekStart.Before(to.UTC())
}

type insightCounts struct {
	events         int
	closed         int
	closedByManual int
	firing         int
}

// buildInsight 合并实时表与归档表的聚合行，并按本地自然日补零。
// started 序列提供新增与级别分布，closed 序列提供闭环数与闭环时长分位数。
func buildInsight(startedHot, startedArchived, closedHot, closedArchived []insightRow, from, to time.Time, loc *time.Location, complete bool) bo.AlertInsight {
	if loc == nil {
		loc = time.UTC
	}
	byDay := map[string]*insightCounts{}
	byLevel := map[string]*insightCounts{}
	eventsTotal := 0
	for _, rows := range [][]insightRow{startedHot, startedArchived} {
		for _, row := range rows {
			if row.day == "" || row.status == "" {
				continue
			}
			day := byDay[row.day]
			if day == nil {
				day = &insightCounts{}
				byDay[row.day] = day
			}
			level := byLevel[row.levelUID]
			if level == nil {
				level = &insightCounts{}
				byLevel[row.levelUID] = level
			}
			day.events += row.events
			level.events += row.events
			eventsTotal += row.events
			if row.status == "FIRING" {
				day.firing += row.events
				level.firing += row.events
			}
		}
	}

	closedTotal := 0
	for _, rows := range [][]insightRow{closedHot, closedArchived} {
		for _, row := range rows {
			if row.day == "" || row.status == "" {
				continue
			}
			day := byDay[row.day]
			if day == nil {
				day = &insightCounts{}
				byDay[row.day] = day
			}
			day.closed += row.events
			closedTotal += row.events
			if row.status == "RECOVERED_BY_MANUAL" {
				day.closedByManual += row.events
			}
		}
	}

	quantileOfDay := func(day string) insightRow {
		// 归档优先：已闭环的事件都在归档里，主库只剩归档失败的兜底行。
		if row, ok := findInsightSummary(closedArchived, day); ok && row.samples > 0 {
			return row
		}
		row, _ := findInsightSummary(closedHot, day)
		return row
	}

	days := make([]bo.AlertInsightDay, 0, 32)
	cursor := from.In(loc)
	for cursor.Before(to) {
		key := cursor.Format("2006-01-02")
		item := bo.AlertInsightDay{Start: cursor}
		if counts := byDay[key]; counts != nil {
			item.Events = counts.events
			item.Closed = counts.closed
			item.ClosedByManual = counts.closedByManual
		}
		if quantile := quantileOfDay(key); quantile.samples > 0 {
			item.P50Seconds = quantile.p50
			item.P90Seconds = quantile.p90
			item.RecoverySamples = quantile.samples
		}
		days = append(days, item)
		cursor = cursor.AddDate(0, 0, 1)
	}

	levels := make([]bo.AlertInsightLevel, 0, len(byLevel))
	for levelUID, counts := range byLevel {
		levels = append(levels, bo.AlertInsightLevel{
			LevelUID: levelUID,
			Events:   counts.events,
			Firing:   counts.firing,
		})
	}
	sort.Slice(levels, func(i, j int) bool {
		if levels[i].Events != levels[j].Events {
			return levels[i].Events > levels[j].Events
		}
		return levels[i].LevelUID < levels[j].LevelUID
	})

	insight := bo.AlertInsight{
		TimeZone:    loc.String(),
		From:        from,
		To:          to,
		Days:        days,
		Levels:      levels,
		EventsTotal: eventsTotal,
		ClosedTotal: closedTotal,
		Complete:    complete,
	}
	if row, ok := findInsightSummary(closedArchived, ""); ok && row.samples > 0 {
		insight.P50Seconds = row.p50
		insight.P90Seconds = row.p90
		insight.RecoverySamples = row.samples
	} else if row, ok := findInsightSummary(closedHot, ""); ok && row.samples > 0 {
		insight.P50Seconds = row.p50
		insight.P90Seconds = row.p90
		insight.RecoverySamples = row.samples
	}
	return insight
}

// findInsightSummary 取出汇总行：day 为空取区间汇总，否则取该日的汇总行
// （status 与 level_uid 都为空的那一行）。
func findInsightSummary(rows []insightRow, day string) (insightRow, bool) {
	for _, row := range rows {
		if row.status != "" || row.levelUID != "" || row.day != day {
			continue
		}
		return row, true
	}
	return insightRow{}, false
}
