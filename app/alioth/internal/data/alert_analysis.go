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

// analysisHistoryPredicate 是历史事件的统计谓词，与历史明细列表的 historyItems 口径一致：
// 已结束（非 FIRING）的事件，或仍在抑制中的 FIRING 事件。
const analysisHistoryPredicate = `(status <> 'FIRING' OR (suppressed = TRUE AND (suppressed_until IS NULL OR suppressed_until > NOW())))`

// analysisSourceColumns 是分析聚合需要的列。窗口与谓词只用到 uid/ends_at/suppressed，
// 其余列供分组与计数。
const analysisSourceColumns = "uid, rule_uid, fingerprint, status, level_uid, starts_at, ends_at"

const (
	// analysisDefaultFingerprintLimit 与 biz 层的默认值一致，仅作数据层兜底。
	analysisDefaultFingerprintLimit = 5
	// analysisDefaultGranularity 是趋势分桶的默认粒度。
	analysisDefaultGranularity = "day"
)

// analysisArgs 按调用顺序收集 SQL 参数，并返回对应的占位符。
// 归档周表的 UNION ALL 分支各自携带租户与窗口参数，用收集器避免手工数占位符。
type analysisArgs struct{ values []any }

func (a *analysisArgs) add(value any) string {
	a.values = append(a.values, value)
	return fmt.Sprintf("$%d", len(a.values))
}

// analysisSource 构造一个来源（主库或某张归档周表）的统计源集合：
// 租户匹配、starts_at 落在窗口内、历史事件谓词，再叠加状态过滤与主库 uid 去重。
func analysisSource(scope bo.AlertAnalysisScope, namespaceUID, relation string, exclude []string, args *analysisArgs) string {
	conditions := []string{
		"namespace_uid = " + args.add(namespaceUID),
		"starts_at >= " + args.add(scope.From),
		"starts_at < " + args.add(scope.To),
		analysisHistoryPredicate,
	}
	if scope.Status != "" {
		// SUPPRESSING 由 biz 层换算成 status = 'FIRING'，这里只按状态值过滤。
		conditions = append(conditions, "status = "+args.add(scope.Status))
	}
	if len(exclude) > 0 {
		conditions = append(conditions, "NOT (uid = ANY("+args.add(exclude)+"))")
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", analysisSourceColumns, relation, strings.Join(conditions, " AND "))
}

// analysisLocation 兜底时区：数据层不假设调用方一定传了 Location。
func analysisLocation(scope bo.AlertAnalysisScope) *time.Location {
	if scope.Location == nil {
		return time.UTC
	}
	return scope.Location
}

// normalizeAnalysisScope 补齐时区，其余字段原样使用。
func normalizeAnalysisScope(scope bo.AlertAnalysisScope) bo.AlertAnalysisScope {
	scope.Location = analysisLocation(scope)
	return scope
}

// analysisGranularity 兜底分桶粒度。周桶由 PostgreSQL 的 date_trunc('week') 提供，
// 即 ISO 周、周一起始。
func analysisGranularity(value string) string {
	switch value {
	case "hour", "day", "week", "month":
		return value
	default:
		return analysisDefaultGranularity
	}
}

// analysisFingerprintSQL 按指纹分组统计事件数。不用 count(DISTINCT fingerprint)：
// 那会退化成外部排序，group by 的 hash 聚合在 25 万行上快一个数量级。
func analysisFingerprintSQL(source string) string {
	return fmt.Sprintf(`SELECT fingerprint, count(*) AS events FROM (%s) e GROUP BY fingerprint`, source)
}

// analysisFingerprintGroupSQL 按指纹分组统计事件数与所属规则。
// 同一指纹由 fingerprint.Of(rule, datasource, labels) 生成，规则唯一，min(rule_uid) 只是
// 让聚合结果确定，不会把不同规则混在一起。
func analysisFingerprintGroupSQL(source string) string {
	return fmt.Sprintf(`SELECT fingerprint, min(rule_uid) AS rule_uid, count(*) AS events
		FROM (%s) e WHERE fingerprint <> '' GROUP BY fingerprint`, source)
}

// analysisRecoverySQL 只对 ends_at 非空的行算分位数：
// 中位数用 percentile_cont(0.5)，P90 用 percentile_disc(0.9)（最近秩口径，
// 与历史分析页原来的 durationSummary 一致）。分析接口里只有这一个查询用分位数。
func analysisRecoverySQL(source string) string {
	return fmt.Sprintf(`SELECT status,
		count(*) FILTER (WHERE ends_at IS NOT NULL) AS samples,
		coalesce(percentile_cont(0.5) WITHIN GROUP (ORDER BY extract(epoch FROM ends_at - starts_at))
			FILTER (WHERE ends_at IS NOT NULL), 0) AS median_seconds,
		coalesce(percentile_disc(0.9) WITHIN GROUP (ORDER BY extract(epoch FROM ends_at - starts_at))
			FILTER (WHERE ends_at IS NOT NULL), 0) AS p90_seconds
	FROM (%s) e
	GROUP BY status`, source)
}

// scanAnalysisKeyCounts 读取 (key, events) 两列。
func scanAnalysisKeyCounts(rows *sql.Rows) (map[string]int64, error) {
	defer func() { _ = rows.Close() }()
	counts := make(map[string]int64)
	for rows.Next() {
		var key string
		var events int64
		if err := rows.Scan(&key, &events); err != nil {
			return nil, err
		}
		counts[key] = events
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

// analysisFingerprintGroup 是按指纹分组的事件数和所属规则。
type analysisFingerprintGroup struct {
	events  int64
	ruleUID string
}

// scanAnalysisFingerprintGroups 读取 (fingerprint, rule_uid, events) 三列。
func scanAnalysisFingerprintGroups(rows *sql.Rows) (map[string]analysisFingerprintGroup, error) {
	defer func() { _ = rows.Close() }()
	groups := make(map[string]analysisFingerprintGroup)
	for rows.Next() {
		var fingerprint, ruleUID string
		var events int64
		if err := rows.Scan(&fingerprint, &ruleUID, &events); err != nil {
			return nil, err
		}
		groups[fingerprint] = analysisFingerprintGroup{events: events, ruleUID: ruleUID}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

// scanAnalysisBucketCounts 读取 (桶起点, 事件数) 两列，用 Unix 秒作为合并键。
func scanAnalysisBucketCounts(rows *sql.Rows) (map[int64]int64, error) {
	defer func() { _ = rows.Close() }()
	counts := make(map[int64]int64)
	for rows.Next() {
		var start time.Time
		var events int64
		if err := rows.Scan(&start, &events); err != nil {
			return nil, err
		}
		counts[start.UTC().Unix()] += events
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

// scanAnalysisRecovery 读取 (status, samples, median, p90) 四列。
func scanAnalysisRecovery(rows *sql.Rows) (map[string]bo.AlertAnalysisRecoveryStats, error) {
	defer func() { _ = rows.Close() }()
	stats := make(map[string]bo.AlertAnalysisRecoveryStats)
	for rows.Next() {
		var (
			status string
			item   bo.AlertAnalysisRecoveryStats
			median sql.NullFloat64
			p90    sql.NullFloat64
		)
		if err := rows.Scan(&status, &item.Samples, &median, &p90); err != nil {
			return nil, err
		}
		item.MedianSeconds = median.Float64
		item.P90Seconds = p90.Float64
		stats[status] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// analysisKeyCount 是排序用的分组计数。
type analysisKeyCount struct {
	key    string
	events int64
}

// mergeAnalysisCounts 合并主库与归档的分组计数。主库与归档已按 uid 去重，
// 同一分组的计数可以直接相加。
func mergeAnalysisCounts(parts ...map[string]int64) map[string]int64 {
	merged := make(map[string]int64)
	for _, part := range parts {
		for key, events := range part {
			merged[key] += events
		}
	}
	return merged
}

// sortedAnalysisKeyCounts 按事件数降序、键升序返回分组计数，保证接口输出确定。
func sortedAnalysisKeyCounts(counts map[string]int64) []analysisKeyCount {
	items := make([]analysisKeyCount, 0, len(counts))
	for key, events := range counts {
		items = append(items, analysisKeyCount{key: key, events: events})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].events != items[j].events {
			return items[i].events > items[j].events
		}
		return items[i].key < items[j].key
	})
	return items
}

// analysisFingerprintTotals 汇总按指纹分组的事件数：
// 事件总数、去重指纹数（空指纹不计入）和重复事件数（同一指纹第 2 次及以后的出现次数）。
func analysisFingerprintTotals(groups map[string]int64) (events, unique, repeated int64) {
	for fingerprint, count := range groups {
		events += count
		if fingerprint == "" {
			continue
		}
		unique++
		repeated += count - 1
	}
	return events, unique, repeated
}

// topAnalysisFingerprints 只保留出现 2 次及以上的指纹，按事件数降序、指纹升序取前 limit 个。
func topAnalysisFingerprints(groups map[string]analysisFingerprintGroup, limit int) []bo.AlertAnalysisFingerprint {
	if limit <= 0 {
		limit = analysisDefaultFingerprintLimit
	}
	items := make([]bo.AlertAnalysisFingerprint, 0, len(groups))
	for fingerprint, group := range groups {
		if group.events <= 1 {
			continue
		}
		items = append(items, bo.AlertAnalysisFingerprint{
			Fingerprint: fingerprint,
			RuleUID:     group.ruleUID,
			Events:      group.events,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Events != items[j].Events {
			return items[i].Events > items[j].Events
		}
		return items[i].Fingerprint < items[j].Fingerprint
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

// analysisRecoveryOf 取一类恢复方式的分位数：归档优先，因为已闭环的事件都在归档里，
// 主库只剩归档失败的兜底行；归档没有该类样本时退回主库，避免分位数恒为 0。
func analysisRecoveryOf(archived, hot map[string]bo.AlertAnalysisRecoveryStats, status string) bo.AlertAnalysisRecoveryStats {
	if stats := archived[status]; stats.Samples > 0 {
		return stats
	}
	return hot[status]
}

// analysisHistory 在租户历史库覆盖窗口的周表上执行一次聚合查询。
// build 接收周表的 UNION ALL 源集合与参数收集器，scan 读取结果。
// 没有租户库或没有覆盖窗口的周表时不执行查询；查询或扫描失败时返回
// 「历史存储不可用」，由调用方降级为只统计实时表。
func (s *historyStore) analysisHistory(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope, exclude []string,
	build func(source string, args *analysisArgs) string, scan func(rows *sql.Rows) error) error {
	conn, tables, err := s.historyTables(ctx, namespaceUID, scope.From.Add(-insightTableCushion), scope.To)
	if err != nil {
		return err
	}
	if len(tables) == 0 {
		return nil
	}
	args := &analysisArgs{}
	branches := make([]string, 0, len(tables))
	for _, table := range tables {
		branches = append(branches, analysisSource(scope, namespaceUID, quoteShardIdent(table), exclude, args))
	}
	rows, err := conn.QueryContext(ctx, build(strings.Join(branches, "\nUNION ALL\n"), args), args.values...)
	if err != nil {
		s.log.Error(fmt.Sprintf("history analysis %s: %v", s.databaseName(namespaceUID), err))
		return merr.Server("history storage is unavailable")
	}
	if err := scan(rows); err != nil {
		s.log.Error(fmt.Sprintf("history analysis scan %s: %v", s.databaseName(namespaceUID), err))
		return merr.Server("history storage is unavailable")
	}
	return nil
}

// analysisHistoryExclude 取主库窗口里「本次统计会算到」的 uid，供归档查询排除重复行。
// 去重集合必须和统计集合用同一个谓词：主库里仍是 FIRING、但归档里已闭环的同一 uid
// （归档后主库行没删掉、又重新触发的场景）不算主库侧的统计行，归档记录仍要计数，
// 否则历史分析会比历史明细少掉这些事件。主库行确实参与统计时仍由主库优先。
func (r *alertEventRepo) analysisHistoryExclude(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) ([]string, error) {
	args := &analysisArgs{}
	source := analysisSource(scope, namespaceUID, "alert_events", nil, args)
	query := fmt.Sprintf("SELECT uid FROM (%s) e LIMIT %d", source, insightHotUIDLimit)
	rows, err := r.db.QueryContext(ctx, query, args.values...)
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

// AnalysisSummary 统计窗口内的事件总数、去重指纹数和重复事件数。
func (r *alertEventRepo) AnalysisSummary(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (bo.AlertAnalysisSummary, error) {
	scope = normalizeAnalysisScope(scope)
	hot, err := r.analysisFingerprintCounts(ctx, namespaceUID, scope)
	if err != nil {
		return bo.AlertAnalysisSummary{}, err
	}
	archived := map[string]int64{}
	complete := false
	if r.history != nil {
		exclude, err := r.analysisHistoryExclude(ctx, namespaceUID, scope)
		if err != nil {
			return bo.AlertAnalysisSummary{}, err
		}
		err = r.history.analysisHistory(ctx, namespaceUID, scope, exclude,
			func(source string, _ *analysisArgs) string { return analysisFingerprintSQL(source) },
			func(rows *sql.Rows) (err error) {
				archived, err = scanAnalysisKeyCounts(rows)
				return err
			})
		if err != nil {
			archived = map[string]int64{}
		} else {
			complete = true
		}
	}
	byFingerprint := mergeAnalysisCounts(hot, archived)
	events, unique, repeated := analysisFingerprintTotals(byFingerprint)
	loc := analysisLocation(scope)
	return bo.AlertAnalysisSummary{
		TimeZone:           loc.String(),
		From:               scope.From.In(loc),
		To:                 scope.To.In(loc),
		Complete:           complete,
		Events:             events,
		UniqueFingerprints: unique,
		Repeated:           repeated,
	}, nil
}

// analysisFingerprintCounts 在实时表上按指纹分组统计事件数。
func (r *alertEventRepo) analysisFingerprintCounts(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (map[string]int64, error) {
	args := &analysisArgs{}
	source := analysisSource(scope, namespaceUID, "alert_events", nil, args)
	rows, err := r.db.QueryContext(ctx, analysisFingerprintSQL(source), args.values...)
	if err != nil {
		return nil, r.fail(err)
	}
	counts, err := scanAnalysisKeyCounts(rows)
	if err != nil {
		return nil, r.fail(err)
	}
	return counts, nil
}

// AnalysisTrend 按本地时区的小时、天、周或月分桶统计事件数，只返回非空桶。
func (r *alertEventRepo) AnalysisTrend(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope, granularity string) (bo.AlertAnalysisTrend, error) {
	scope = normalizeAnalysisScope(scope)
	granularity = analysisGranularity(granularity)
	loc := analysisLocation(scope)
	hot, err := r.analysisBuckets(ctx, namespaceUID, scope, granularity)
	if err != nil {
		return bo.AlertAnalysisTrend{}, err
	}
	archived := map[int64]int64{}
	complete := false
	if r.history != nil {
		exclude, err := r.analysisHistoryExclude(ctx, namespaceUID, scope)
		if err != nil {
			return bo.AlertAnalysisTrend{}, err
		}
		err = r.history.analysisHistory(ctx, namespaceUID, scope, exclude,
			func(source string, args *analysisArgs) string {
				tz := args.add(loc.String())
				return fmt.Sprintf(`SELECT date_trunc('%s', starts_at AT TIME ZONE %s) AT TIME ZONE %s AS bucket, count(*) AS events
					FROM (%s) e GROUP BY 1`, granularity, tz, tz, source)
			},
			func(rows *sql.Rows) (err error) {
				archived, err = scanAnalysisBucketCounts(rows)
				return err
			})
		if err != nil {
			archived = map[int64]int64{}
		} else {
			complete = true
		}
	}
	counts := mergeAnalysisBucketCounts(hot, archived)
	keys := make([]int64, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	buckets := make([]bo.AlertAnalysisBucket, 0, len(keys))
	for _, key := range keys {
		buckets = append(buckets, bo.AlertAnalysisBucket{Start: time.Unix(key, 0).In(loc), Events: counts[key]})
	}
	return bo.AlertAnalysisTrend{
		TimeZone: loc.String(),
		From:     scope.From.In(loc),
		To:       scope.To.In(loc),
		Complete: complete,
		Buckets:  buckets,
	}, nil
}

// analysisBuckets 在实时表上按粒度分桶统计。
func (r *alertEventRepo) analysisBuckets(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope, granularity string) (map[int64]int64, error) {
	loc := analysisLocation(scope)
	args := &analysisArgs{}
	source := analysisSource(scope, namespaceUID, "alert_events", nil, args)
	tz := args.add(loc.String())
	query := fmt.Sprintf(`SELECT date_trunc('%s', starts_at AT TIME ZONE %s) AT TIME ZONE %s AS bucket, count(*) AS events
		FROM (%s) e GROUP BY 1`, granularity, tz, tz, source)
	rows, err := r.db.QueryContext(ctx, query, args.values...)
	if err != nil {
		return nil, r.fail(err)
	}
	counts, err := scanAnalysisBucketCounts(rows)
	if err != nil {
		return nil, r.fail(err)
	}
	return counts, nil
}

// mergeAnalysisBucketCounts 合并主库与归档的分桶计数。
func mergeAnalysisBucketCounts(parts ...map[int64]int64) map[int64]int64 {
	merged := make(map[int64]int64)
	for _, part := range parts {
		for key, events := range part {
			merged[key] += events
		}
	}
	return merged
}

// AnalysisStatus 统计窗口内按事件状态分布的事件数。
func (r *alertEventRepo) AnalysisStatus(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (bo.AlertAnalysisStatus, error) {
	scope = normalizeAnalysisScope(scope)
	hot, err := r.analysisGroupCounts(ctx, namespaceUID, scope, "status")
	if err != nil {
		return bo.AlertAnalysisStatus{}, err
	}
	archived := map[string]int64{}
	complete := false
	if r.history != nil {
		exclude, err := r.analysisHistoryExclude(ctx, namespaceUID, scope)
		if err != nil {
			return bo.AlertAnalysisStatus{}, err
		}
		err = r.history.analysisHistory(ctx, namespaceUID, scope, exclude,
			func(source string, args *analysisArgs) string {
				return fmt.Sprintf(`SELECT status, count(*) AS events FROM (%s) e GROUP BY status`, source)
			},
			func(rows *sql.Rows) (err error) {
				archived, err = scanAnalysisKeyCounts(rows)
				return err
			})
		if err != nil {
			archived = map[string]int64{}
		} else {
			complete = true
		}
	}
	items, total := analysisStatusItems(mergeAnalysisCounts(hot, archived))
	return bo.AlertAnalysisStatus{Complete: complete, Total: total, Items: items}, nil
}

// analysisStatusItems 把状态分组计数转换成按事件数降序的列表。
func analysisStatusItems(counts map[string]int64) ([]bo.AlertAnalysisStatusCount, int64) {
	sorted := sortedAnalysisKeyCounts(counts)
	items := make([]bo.AlertAnalysisStatusCount, 0, len(sorted))
	var total int64
	for _, item := range sorted {
		items = append(items, bo.AlertAnalysisStatusCount{Status: item.key, Events: item.events})
		total += item.events
	}
	return items, total
}

// AnalysisLevel 统计窗口内按告警级别分布的事件数。
func (r *alertEventRepo) AnalysisLevel(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (bo.AlertAnalysisLevel, error) {
	scope = normalizeAnalysisScope(scope)
	hot, err := r.analysisGroupCounts(ctx, namespaceUID, scope, "level_uid")
	if err != nil {
		return bo.AlertAnalysisLevel{}, err
	}
	archived := map[string]int64{}
	complete := false
	if r.history != nil {
		exclude, err := r.analysisHistoryExclude(ctx, namespaceUID, scope)
		if err != nil {
			return bo.AlertAnalysisLevel{}, err
		}
		err = r.history.analysisHistory(ctx, namespaceUID, scope, exclude,
			func(source string, args *analysisArgs) string {
				return fmt.Sprintf(`SELECT level_uid, count(*) AS events FROM (%s) e GROUP BY level_uid`, source)
			},
			func(rows *sql.Rows) (err error) {
				archived, err = scanAnalysisKeyCounts(rows)
				return err
			})
		if err != nil {
			archived = map[string]int64{}
		} else {
			complete = true
		}
	}
	items, total := analysisLevelItems(mergeAnalysisCounts(hot, archived))
	return bo.AlertAnalysisLevel{Complete: complete, Total: total, Items: items}, nil
}

// analysisLevelItems 把级别分组计数转换成按事件数降序的列表。
func analysisLevelItems(counts map[string]int64) ([]bo.AlertAnalysisLevelCount, int64) {
	sorted := sortedAnalysisKeyCounts(counts)
	items := make([]bo.AlertAnalysisLevelCount, 0, len(sorted))
	var total int64
	for _, item := range sorted {
		items = append(items, bo.AlertAnalysisLevelCount{LevelUID: item.key, Events: item.events})
		total += item.events
	}
	return items, total
}

// analysisGroupCounts 在实时表上按列分组计数。column 只由本文件的常量传入。
func (r *alertEventRepo) analysisGroupCounts(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope, column string) (map[string]int64, error) {
	if column != "status" && column != "level_uid" {
		return nil, merr.Server("invalid analysis grouping")
	}
	args := &analysisArgs{}
	source := analysisSource(scope, namespaceUID, "alert_events", nil, args)
	query := fmt.Sprintf(`SELECT %s, count(*) AS events FROM (%s) e GROUP BY %s`, column, source, column)
	rows, err := r.db.QueryContext(ctx, query, args.values...)
	if err != nil {
		return nil, r.fail(err)
	}
	counts, err := scanAnalysisKeyCounts(rows)
	if err != nil {
		return nil, r.fail(err)
	}
	return counts, nil
}

// AnalysisFingerprints 统计出现 2 次及以上的指纹，按事件数降序取前 limit 个。
func (r *alertEventRepo) AnalysisFingerprints(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope, limit int) (bo.AlertAnalysisFingerprints, error) {
	scope = normalizeAnalysisScope(scope)
	if limit <= 0 {
		limit = analysisDefaultFingerprintLimit
	}
	hot, err := r.analysisFingerprintGroups(ctx, namespaceUID, scope)
	if err != nil {
		return bo.AlertAnalysisFingerprints{}, err
	}
	archived := map[string]analysisFingerprintGroup{}
	complete := false
	if r.history != nil {
		exclude, err := r.analysisHistoryExclude(ctx, namespaceUID, scope)
		if err != nil {
			return bo.AlertAnalysisFingerprints{}, err
		}
		err = r.history.analysisHistory(ctx, namespaceUID, scope, exclude,
			func(source string, _ *analysisArgs) string { return analysisFingerprintGroupSQL(source) },
			func(rows *sql.Rows) (err error) {
				archived, err = scanAnalysisFingerprintGroups(rows)
				return err
			})
		if err != nil {
			archived = map[string]analysisFingerprintGroup{}
		} else {
			complete = true
		}
	}
	items := topAnalysisFingerprints(mergeAnalysisFingerprintGroups(hot, archived), limit)
	ruleUIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item.RuleUID != "" {
			ruleUIDs = append(ruleUIDs, item.RuleUID)
		}
	}
	names := r.ruleNames(ctx, ruleUIDs)
	for index := range items {
		items[index].RuleName = names[items[index].RuleUID].Name
	}
	return bo.AlertAnalysisFingerprints{Complete: complete, Items: items}, nil
}

// analysisFingerprintGroups 在实时表上按指纹分组取事件数与规则标识。
func (r *alertEventRepo) analysisFingerprintGroups(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (map[string]analysisFingerprintGroup, error) {
	args := &analysisArgs{}
	source := analysisSource(scope, namespaceUID, "alert_events", nil, args)
	rows, err := r.db.QueryContext(ctx, analysisFingerprintGroupSQL(source), args.values...)
	if err != nil {
		return nil, r.fail(err)
	}
	groups, err := scanAnalysisFingerprintGroups(rows)
	if err != nil {
		return nil, r.fail(err)
	}
	return groups, nil
}

// mergeAnalysisFingerprintGroups 合并主库与归档的指纹分组，主库的规则标识优先。
func mergeAnalysisFingerprintGroups(parts ...map[string]analysisFingerprintGroup) map[string]analysisFingerprintGroup {
	merged := make(map[string]analysisFingerprintGroup)
	for _, part := range parts {
		for fingerprint, group := range part {
			current := merged[fingerprint]
			current.events += group.events
			if current.ruleUID == "" {
				current.ruleUID = group.ruleUID
			}
			merged[fingerprint] = current
		}
	}
	return merged
}

// AnalysisRecovery 统计自动恢复与人工恢复的闭环时长中位数和 P90。
// 分位数由归档优先提供，归档不可用时降级为主库结果并把 Complete 置为 false。
func (r *alertEventRepo) AnalysisRecovery(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (bo.AlertAnalysisRecovery, error) {
	scope = normalizeAnalysisScope(scope)
	hot, err := r.analysisRecoveryStats(ctx, namespaceUID, scope)
	if err != nil {
		return bo.AlertAnalysisRecovery{}, err
	}
	archived := map[string]bo.AlertAnalysisRecoveryStats{}
	complete := false
	if r.history != nil {
		exclude, err := r.analysisHistoryExclude(ctx, namespaceUID, scope)
		if err != nil {
			return bo.AlertAnalysisRecovery{}, err
		}
		err = r.history.analysisHistory(ctx, namespaceUID, scope, exclude,
			func(source string, _ *analysisArgs) string { return analysisRecoverySQL(source) },
			func(rows *sql.Rows) (err error) {
				archived, err = scanAnalysisRecovery(rows)
				return err
			})
		if err != nil {
			archived = map[string]bo.AlertAnalysisRecoveryStats{}
		} else {
			complete = true
		}
	}
	return bo.AlertAnalysisRecovery{
		Complete: complete,
		Auto:     analysisRecoveryOf(archived, hot, "RECOVERED"),
		Manual:   analysisRecoveryOf(archived, hot, "RECOVERED_BY_MANUAL"),
	}, nil
}

// analysisRecoveryStats 在实时表上按恢复方式统计闭环时长。
func (r *alertEventRepo) analysisRecoveryStats(ctx context.Context, namespaceUID string, scope bo.AlertAnalysisScope) (map[string]bo.AlertAnalysisRecoveryStats, error) {
	args := &analysisArgs{}
	source := analysisSource(scope, namespaceUID, "alert_events", nil, args)
	rows, err := r.db.QueryContext(ctx, analysisRecoverySQL(source), args.values...)
	if err != nil {
		return nil, r.fail(err)
	}
	stats, err := scanAnalysisRecovery(rows)
	if err != nil {
		return nil, r.fail(err)
	}
	return stats, nil
}
