package alertevent

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// AnalysisDefaultFingerprintLimit 是重复指纹接口未传 limit 时的默认条数。
	AnalysisDefaultFingerprintLimit int32 = 5
	// AnalysisMaxFingerprintLimit 是重复指纹接口允许的最大条数。
	AnalysisMaxFingerprintLimit int32 = 20
	// analysisDefaultGranularity 是趋势接口未传粒度时的默认值。
	analysisDefaultGranularity = "day"
)

// AnalysisQuery 是历史告警分析接口的查询条件。
// from/to 是本地自然日（含首尾）；focusStart/focusEnd 是可选的聚焦窗口。
type AnalysisQuery struct {
	From        string
	To          string
	TimeZone    string
	Status      string
	FocusStart  string
	FocusEnd    string
	Granularity string
	Limit       int32
}

// AnalysisSummary 返回分析页顶部的汇总指标：事件总数、去重指纹数和重复事件数。
func (coordinator *Coordinator) AnalysisSummary(ctx context.Context, query AnalysisQuery) (bo.AlertAnalysisSummary, error) {
	ns, err := coordinator.access.AllowAny(ctx, perm.AlertRealtimeRead, perm.AlertHistoryRead)
	if err != nil {
		return bo.AlertAnalysisSummary{}, err
	}
	scope, err := analysisScope(query, time.Now())
	if err != nil {
		return bo.AlertAnalysisSummary{}, err
	}
	summary, err := coordinator.events.AnalysisSummary(ctx, ns, scope)
	return summary, coordinator.access.Wrap(err)
}

// AnalysisTrend 返回分析页的趋势分桶。趋势只按日窗口分桶，忽略聚焦窗口。
func (coordinator *Coordinator) AnalysisTrend(ctx context.Context, query AnalysisQuery) (bo.AlertAnalysisTrend, error) {
	ns, err := coordinator.access.AllowAny(ctx, perm.AlertRealtimeRead, perm.AlertHistoryRead)
	if err != nil {
		return bo.AlertAnalysisTrend{}, err
	}
	// 聚焦是统计图表的交互，趋势图不参与：时间桶本身就是它的聚焦维度。
	query.FocusStart = ""
	query.FocusEnd = ""
	scope, err := analysisScope(query, time.Now())
	if err != nil {
		return bo.AlertAnalysisTrend{}, err
	}
	granularity, err := analysisGranularity(query.Granularity)
	if err != nil {
		return bo.AlertAnalysisTrend{}, err
	}
	trend, err := coordinator.events.AnalysisTrend(ctx, ns, scope, granularity)
	return trend, coordinator.access.Wrap(err)
}

// AnalysisStatus 返回分析页的事件状态分布。
func (coordinator *Coordinator) AnalysisStatus(ctx context.Context, query AnalysisQuery) (bo.AlertAnalysisStatus, error) {
	ns, err := coordinator.access.AllowAny(ctx, perm.AlertRealtimeRead, perm.AlertHistoryRead)
	if err != nil {
		return bo.AlertAnalysisStatus{}, err
	}
	scope, err := analysisScope(query, time.Now())
	if err != nil {
		return bo.AlertAnalysisStatus{}, err
	}
	status, err := coordinator.events.AnalysisStatus(ctx, ns, scope)
	return status, coordinator.access.Wrap(err)
}

// AnalysisLevel 返回分析页的告警级别分布。
func (coordinator *Coordinator) AnalysisLevel(ctx context.Context, query AnalysisQuery) (bo.AlertAnalysisLevel, error) {
	ns, err := coordinator.access.AllowAny(ctx, perm.AlertRealtimeRead, perm.AlertHistoryRead)
	if err != nil {
		return bo.AlertAnalysisLevel{}, err
	}
	scope, err := analysisScope(query, time.Now())
	if err != nil {
		return bo.AlertAnalysisLevel{}, err
	}
	level, err := coordinator.events.AnalysisLevel(ctx, ns, scope)
	return level, coordinator.access.Wrap(err)
}

// AnalysisFingerprints 返回出现 2 次及以上的指纹，按事件数降序取前 limit 个。
func (coordinator *Coordinator) AnalysisFingerprints(ctx context.Context, query AnalysisQuery) (bo.AlertAnalysisFingerprints, error) {
	ns, err := coordinator.access.AllowAny(ctx, perm.AlertRealtimeRead, perm.AlertHistoryRead)
	if err != nil {
		return bo.AlertAnalysisFingerprints{}, err
	}
	scope, err := analysisScope(query, time.Now())
	if err != nil {
		return bo.AlertAnalysisFingerprints{}, err
	}
	limit, err := analysisFingerprintLimit(query.Limit)
	if err != nil {
		return bo.AlertAnalysisFingerprints{}, err
	}
	fingerprints, err := coordinator.events.AnalysisFingerprints(ctx, ns, scope, limit)
	return fingerprints, coordinator.access.Wrap(err)
}

// AnalysisRecovery 返回自动恢复与人工恢复的闭环时长中位数和 P90。
func (coordinator *Coordinator) AnalysisRecovery(ctx context.Context, query AnalysisQuery) (bo.AlertAnalysisRecovery, error) {
	ns, err := coordinator.access.AllowAny(ctx, perm.AlertRealtimeRead, perm.AlertHistoryRead)
	if err != nil {
		return bo.AlertAnalysisRecovery{}, err
	}
	scope, err := analysisScope(query, time.Now())
	if err != nil {
		return bo.AlertAnalysisRecovery{}, err
	}
	recovery, err := coordinator.events.AnalysisRecovery(ctx, ns, scope)
	return recovery, coordinator.access.Wrap(err)
}

// analysisScope 把分析接口的日期、时区、状态和聚焦参数换算成聚合窗口。
// 校验失败一律返回 merr.Invalid，HTTP 层映射为 400。
func analysisScope(query AnalysisQuery, now time.Time) (bo.AlertAnalysisScope, error) {
	loc, err := analysisTimeZone(query.TimeZone)
	if err != nil {
		return bo.AlertAnalysisScope{}, err
	}
	from, to, err := analysisWindow(query.From, query.To, loc, now)
	if err != nil {
		return bo.AlertAnalysisScope{}, err
	}
	status, err := analysisStatus(query.Status)
	if err != nil {
		return bo.AlertAnalysisScope{}, err
	}
	if status == "SUPPRESSING" {
		// 仍被抑制的事件在库里仍是 FIRING，由历史事件谓词保证「仍在抑制中」。
		status = "FIRING"
	}
	from, to, err = analysisFocus(from, to, query.FocusStart, query.FocusEnd)
	if err != nil {
		return bo.AlertAnalysisScope{}, err
	}
	return bo.AlertAnalysisScope{From: from, To: to, Location: loc, Status: status}, nil
}

// analysisTimeZone 解析 IANA 时区，空值按 UTC。
func analysisTimeZone(value string) (*time.Location, error) {
	if value == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(value)
	if err != nil {
		return nil, merr.Invalid("invalid time zone")
	}
	return loc, nil
}

// analysisWindow 把 YYYY-MM-DD 的本地自然日换算成 [当天零点, 结束日次日零点) 的窗口。
// to 不得晚于时区下的今天，跨度最多 MaxHistoryRangeDays 天（含首尾）。
func analysisWindow(fromValue, toValue string, loc *time.Location, now time.Time) (time.Time, time.Time, error) {
	if fromValue == "" || toValue == "" {
		return time.Time{}, time.Time{}, merr.Invalid("invalid request")
	}
	from, err := time.ParseInLocation("2006-01-02", fromValue, loc)
	if err != nil {
		return time.Time{}, time.Time{}, merr.Invalid("invalid request")
	}
	to, err := time.ParseInLocation("2006-01-02", toValue, loc)
	if err != nil {
		return time.Time{}, time.Time{}, merr.Invalid("invalid request")
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, merr.Invalid("invalid request")
	}
	// 用日历日而不是小时数算跨度：夏令时切换日不足 24 小时，不能让 30 天变成 29 天。
	fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	if days := int(toDay.Sub(fromDay)/(24*time.Hour)) + 1; days > MaxHistoryRangeDays {
		return time.Time{}, time.Time{}, merr.Invalid("alert analysis range must be within 30 days")
	}
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
	if toDay.After(today) {
		return time.Time{}, time.Time{}, merr.Invalid("alert analysis range must not end in the future")
	}
	return from, to.AddDate(0, 0, 1), nil
}

// analysisStatus 校验状态过滤。SUPPRESSING 由历史事件谓词配合 status = 'FIRING' 表达。
func analysisStatus(value string) (string, error) {
	switch value {
	case "", "RECOVERED", "RECOVERED_BY_MANUAL", "SUPPRESSING":
		return value, nil
	default:
		return "", merr.Invalid("invalid alert analysis status")
	}
}

// analysisFocus 把可选聚焦窗口叠加到日窗口上。两个参数必须成对出现，
// 聚焦窗口要落在 [from, to) 内；未传时保持整段统计。
func analysisFocus(from, to time.Time, focusStartValue, focusEndValue string) (time.Time, time.Time, error) {
	if focusStartValue == "" && focusEndValue == "" {
		return from, to, nil
	}
	if focusStartValue == "" || focusEndValue == "" {
		return time.Time{}, time.Time{}, merr.Invalid("invalid focus window")
	}
	focusStart, err := time.Parse(time.RFC3339Nano, focusStartValue)
	if err != nil {
		return time.Time{}, time.Time{}, merr.Invalid("invalid focus window")
	}
	focusEnd, err := time.Parse(time.RFC3339Nano, focusEndValue)
	if err != nil {
		return time.Time{}, time.Time{}, merr.Invalid("invalid focus window")
	}
	if !focusEnd.After(focusStart) {
		return time.Time{}, time.Time{}, merr.Invalid("invalid focus window")
	}
	if focusStart.Before(from) || focusEnd.After(to) {
		return time.Time{}, time.Time{}, merr.Invalid("focus window must be within the analysis range")
	}
	return focusStart, focusEnd, nil
}

// analysisGranularity 校验趋势分桶粒度，空值按天。
func analysisGranularity(value string) (string, error) {
	switch value {
	case "":
		return analysisDefaultGranularity, nil
	case "hour", "day", "week", "month":
		return value, nil
	default:
		return "", merr.Invalid("invalid alert analysis granularity")
	}
}

// analysisFingerprintLimit 校验重复指纹返回条数，0 表示默认 5，范围 1..20。
func analysisFingerprintLimit(value int32) (int, error) {
	if value == 0 {
		return int(AnalysisDefaultFingerprintLimit), nil
	}
	if value < 1 || value > AnalysisMaxFingerprintLimit {
		return 0, merr.Invalid("invalid fingerprint limit")
	}
	return int(value), nil
}
