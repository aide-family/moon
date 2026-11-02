package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probesladaily"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	labelsutil "github.com/aide-family/moon/pkg/labels"
)

// rawRetention 约为原始探测结果的保留窗口（7 天）。超过这个窗口的 P95
// 无法从原始结果精确计算，报表用日聚合的最大延迟近似并标注。
const rawRetention = 7 * 24 * time.Hour

// Rollup 按天重算 from 之后各自然日的探测聚合。聚合先于原始结果清理执行，
// 写入按 (target_uid, day) 唯一键 upsert，可重复执行。
func (r *probeSlaRepo) Rollup(ctx context.Context, from time.Time) error {
	start := utcDay(from)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO probe_sla_dailies (uid, namespace_uid, group_uid, target_uid, day, checks, ok_checks, failed_checks, total_duration_ms, max_duration_ms, created_at, updated_at)
SELECT gen_random_uuid()::text, namespace_uid, group_uid, target_uid, date_trunc('day', probed_at), count(*),
       count(*) FILTER (WHERE ok), count(*) FILTER (WHERE NOT ok),
       COALESCE(sum(duration_ms), 0), COALESCE(max(duration_ms), 0), now(), now()
FROM probe_results
WHERE probed_at >= $1
GROUP BY namespace_uid, group_uid, target_uid, date_trunc('day', probed_at)
ON CONFLICT (target_uid, day) DO UPDATE SET
  checks = EXCLUDED.checks,
  ok_checks = EXCLUDED.ok_checks,
  failed_checks = EXCLUDED.failed_checks,
  total_duration_ms = EXCLUDED.total_duration_ms,
  max_duration_ms = EXCLUDED.max_duration_ms,
  updated_at = now()`, start)
	return r.fail(err)
}

// Report 汇总目标组的可用率、延迟与事件数，并给出按天明细。
func (r *probeSlaRepo) Report(ctx context.Context, namespaceUID string, groupUIDs []string, from, to time.Time) (bo.ProbeSlaReport, error) {
	report := bo.ProbeSlaReport{From: from, To: to}
	groups, err := r.client.ProbeTargetGroup.Query().Where(
		probetargetgroup.NamespaceUIDEQ(namespaceUID),
		probetargetgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return report, r.fail(err)
	}
	allowed := make(map[string]struct{}, len(groupUIDs))
	for _, uid := range groupUIDs {
		allowed[uid] = struct{}{}
	}
	daily, err := r.client.ProbeSlaDaily.Query().Where(
		probesladaily.NamespaceUIDEQ(namespaceUID),
		probesladaily.DayGTE(utcDay(from)),
		probesladaily.DayLTE(utcDay(to)),
	).All(ctx)
	if err != nil {
		return report, r.fail(err)
	}
	for _, group := range groups {
		if len(allowed) > 0 {
			if _, ok := allowed[group.UID]; !ok {
				continue
			}
		}
		item := bo.ProbeSlaGroup{GroupUID: group.UID, GroupName: group.Name}
		dailyByDay := map[string]*bo.ProbeSlaDay{}
		for _, row := range daily {
			if row.GroupUID != group.UID {
				continue
			}
			item.Checks += row.Checks
			item.OkChecks += row.OkChecks
			item.FailedChecks += row.FailedChecks
			day := row.Day.UTC().Format("2006-01-02")
			entry, ok := dailyByDay[day]
			if !ok {
				entry = &bo.ProbeSlaDay{Day: day}
				dailyByDay[day] = entry
			}
			entry.Checks += row.Checks
			entry.OkChecks += row.OkChecks
			entry.FailedChecks += row.FailedChecks
			entry.AvgDurationMS += int(row.TotalDurationMs)
		}
		item.Availability = formatAvailability(item.OkChecks, item.Checks)
		if item.Checks > 0 {
			item.AvgDurationMS = int(dailyAvgDuration(daily, group.UID) / int64(item.Checks))
		}
		item.P95DurationMS, item.P95Approximate = r.p95Duration(ctx, namespaceUID, group.UID, from, to)
		item.EventCount = r.groupEventCount(ctx, namespaceUID, group.Name, from, to)
		entries := make([]bo.ProbeSlaDay, 0, len(dailyByDay))
		for _, entry := range dailyByDay {
			entry.Availability = formatAvailability(entry.OkChecks, entry.Checks)
			if entry.Checks > 0 {
				entry.AvgDurationMS /= entry.Checks
			}
			entries = append(entries, *entry)
		}
		sortProbeSlaDays(entries)
		item.Days = entries
		report.Groups = append(report.Groups, item)
	}
	return report, nil
}

// dailyAvgDuration 返回组内全部日聚合的延迟总和（毫秒）。
func dailyAvgDuration(rows []*ent.ProbeSlaDaily, groupUID string) int64 {
	var total int64
	for _, row := range rows {
		if row.GroupUID == groupUID {
			total += row.TotalDurationMs
		}
	}
	return total
}

// p95Duration 在原始结果保留期内用 SQL 精确计算 P95；超出保留期时用日聚合的
// 最大延迟近似并标注，避免把近似值冒充分位数。
func (r *probeSlaRepo) p95Duration(ctx context.Context, namespaceUID, groupUID string, from, to time.Time) (int, bool) {
	if time.Since(from) <= rawRetention {
		var p95 *float64
		err := r.db.QueryRowContext(ctx, `
SELECT percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms)
FROM probe_results
WHERE namespace_uid = $1 AND group_uid = $2 AND probed_at >= $3 AND probed_at <= $4`,
			namespaceUID, groupUID, from, to).Scan(&p95)
		if err == nil && p95 != nil {
			return int(*p95), false
		}
	}
	var maxDuration *int64
	err := r.db.QueryRowContext(ctx, `
SELECT max(max_duration_ms) FROM probe_sla_dailies
WHERE namespace_uid = $1 AND group_uid = $2 AND day >= $3 AND day <= $4`,
		namespaceUID, groupUID, utcDay(from), utcDay(to)).Scan(&maxDuration)
	if err != nil || maxDuration == nil {
		return 0, true
	}
	return int(*maxDuration), true
}

// groupEventCount 统计区间内命中该目标组的探测事件数。
func (r *probeSlaRepo) groupEventCount(ctx context.Context, namespaceUID, groupName string, from, to time.Time) int {
	rows, err := r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.LabelsContains("probe_kind"),
		alertevent.StartsAtGTE(from),
		alertevent.StartsAtLTE(to),
	).Select(alertevent.FieldLabels).All(ctx)
	if err != nil {
		return 0
	}
	count := 0
	for _, row := range rows {
		labels, err := labelsutil.ParseObject(row.Labels)
		if err != nil {
			continue
		}
		if labels[datasourceLabelKey] == groupName {
			count++
		}
	}
	return count
}

func sortProbeSlaDays(days []bo.ProbeSlaDay) {
	for i := 1; i < len(days); i++ {
		for j := i; j > 0 && days[j].Day < days[j-1].Day; j-- {
			days[j], days[j-1] = days[j-1], days[j]
		}
	}
}
