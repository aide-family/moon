package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

type AlertEventService struct {
	v1.UnimplementedAlertEventServiceServer
	coordinator *alertevent.Coordinator
}

func NewAlertEventService(coordinator *alertevent.Coordinator) *AlertEventService {
	return &AlertEventService{coordinator: coordinator}
}

func (s *AlertEventService) ListAlertEvents(ctx context.Context, req *v1.ListAlertEventsRequest) (*v1.ListAlertEventsResponse, error) {
	startsAtFrom, err := parseAlertEventTime(req.GetStartsAtFrom())
	if err != nil {
		return nil, err
	}
	startsAtTo, err := parseAlertEventTime(req.GetStartsAtTo())
	if err != nil {
		return nil, err
	}
	items, total, page, pageSize, err := s.coordinator.List(ctx, alertevent.ListQuery{
		View: req.GetView(), Name: req.GetName(), Mine: req.GetMine(),
		PageUID: req.GetPageUid(), LevelUID: req.GetLevelUid(),
		RuleName: req.GetRuleName(), DatasourceName: req.GetDatasourceName(),
		StartsAtFrom: startsAtFrom, StartsAtTo: startsAtTo,
		EventStatus: req.GetEventStatus(), Suppression: req.GetSuppression(),
	}, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertEventsResponse{Items: eventItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *AlertEventService) GetAlertInsight(ctx context.Context, req *v1.GetAlertInsightRequest) (*v1.GetAlertInsightResponse, error) {
	insight, err := s.coordinator.Insight(ctx, req.GetDays(), req.GetTimeZone())
	if err != nil {
		return nil, err
	}
	buckets := make([]*v1.AlertInsightBucket, 0, len(insight.Days))
	for _, day := range insight.Days {
		buckets = append(buckets, &v1.AlertInsightBucket{
			Start:           day.Start.Format(time.RFC3339),
			Events:          int32(day.Events),
			Closed:          int32(day.Closed),
			ClosedByManual:  int32(day.ClosedByManual),
			P50Seconds:      day.P50Seconds,
			P90Seconds:      day.P90Seconds,
			RecoverySamples: int32(day.RecoverySamples),
		})
	}
	levels := make([]*v1.AlertInsightLevel, 0, len(insight.Levels))
	for _, level := range insight.Levels {
		levels = append(levels, &v1.AlertInsightLevel{
			LevelUid: level.LevelUID,
			Events:   int32(level.Events),
			Firing:   int32(level.Firing),
		})
	}
	return &v1.GetAlertInsightResponse{
		TimeZone:        insight.TimeZone,
		From:            insight.From.Format(time.RFC3339),
		To:              insight.To.Format(time.RFC3339),
		Buckets:         buckets,
		Levels:          levels,
		P50Seconds:      insight.P50Seconds,
		P90Seconds:      insight.P90Seconds,
		RecoverySamples: int32(insight.RecoverySamples),
		EventsTotal:     int32(insight.EventsTotal),
		ClosedTotal:     int32(insight.ClosedTotal),
		Complete:        insight.Complete,
	}, nil
}

func (s *AlertEventService) GetAlertAnalysisSummary(ctx context.Context, req *v1.GetAlertAnalysisSummaryRequest) (*v1.GetAlertAnalysisSummaryResponse, error) {
	summary, err := s.coordinator.AnalysisSummary(ctx, alertevent.AnalysisQuery{
		From: req.GetFrom(), To: req.GetTo(), TimeZone: req.GetTimeZone(), Status: req.GetStatus(),
		FocusStart: req.GetFocusStart(), FocusEnd: req.GetFocusEnd(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertAnalysisSummaryResponse{
		From:               summary.From.Format(time.RFC3339),
		To:                 summary.To.Format(time.RFC3339),
		TimeZone:           summary.TimeZone,
		Complete:           summary.Complete,
		Events:             summary.Events,
		UniqueFingerprints: summary.UniqueFingerprints,
		Repeated:           summary.Repeated,
	}, nil
}

func (s *AlertEventService) GetAlertAnalysisTrend(ctx context.Context, req *v1.GetAlertAnalysisTrendRequest) (*v1.GetAlertAnalysisTrendResponse, error) {
	trend, err := s.coordinator.AnalysisTrend(ctx, alertevent.AnalysisQuery{
		From: req.GetFrom(), To: req.GetTo(), TimeZone: req.GetTimeZone(), Status: req.GetStatus(),
		Granularity: req.GetGranularity(),
	})
	if err != nil {
		return nil, err
	}
	buckets := make([]*v1.AlertAnalysisBucket, 0, len(trend.Buckets))
	for _, bucket := range trend.Buckets {
		buckets = append(buckets, &v1.AlertAnalysisBucket{Start: bucket.Start.Format(time.RFC3339), Events: bucket.Events})
	}
	return &v1.GetAlertAnalysisTrendResponse{
		From: trend.From.Format(time.RFC3339), To: trend.To.Format(time.RFC3339),
		TimeZone: trend.TimeZone, Complete: trend.Complete, Buckets: buckets,
	}, nil
}

func (s *AlertEventService) GetAlertAnalysisStatus(ctx context.Context, req *v1.GetAlertAnalysisStatusRequest) (*v1.GetAlertAnalysisStatusResponse, error) {
	status, err := s.coordinator.AnalysisStatus(ctx, alertevent.AnalysisQuery{
		From: req.GetFrom(), To: req.GetTo(), TimeZone: req.GetTimeZone(), Status: req.GetStatus(),
		FocusStart: req.GetFocusStart(), FocusEnd: req.GetFocusEnd(),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AlertAnalysisStatusItem, 0, len(status.Items))
	for _, item := range status.Items {
		items = append(items, &v1.AlertAnalysisStatusItem{Status: item.Status, Events: item.Events})
	}
	return &v1.GetAlertAnalysisStatusResponse{Complete: status.Complete, Total: status.Total, Items: items}, nil
}

func (s *AlertEventService) GetAlertAnalysisLevel(ctx context.Context, req *v1.GetAlertAnalysisLevelRequest) (*v1.GetAlertAnalysisLevelResponse, error) {
	level, err := s.coordinator.AnalysisLevel(ctx, alertevent.AnalysisQuery{
		From: req.GetFrom(), To: req.GetTo(), TimeZone: req.GetTimeZone(), Status: req.GetStatus(),
		FocusStart: req.GetFocusStart(), FocusEnd: req.GetFocusEnd(),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AlertAnalysisLevelItem, 0, len(level.Items))
	for _, item := range level.Items {
		items = append(items, &v1.AlertAnalysisLevelItem{LevelUid: item.LevelUID, Events: item.Events})
	}
	return &v1.GetAlertAnalysisLevelResponse{Complete: level.Complete, Total: level.Total, Items: items}, nil
}

func (s *AlertEventService) GetAlertAnalysisFingerprints(ctx context.Context, req *v1.GetAlertAnalysisFingerprintsRequest) (*v1.GetAlertAnalysisFingerprintsResponse, error) {
	fingerprints, err := s.coordinator.AnalysisFingerprints(ctx, alertevent.AnalysisQuery{
		From: req.GetFrom(), To: req.GetTo(), TimeZone: req.GetTimeZone(), Status: req.GetStatus(),
		FocusStart: req.GetFocusStart(), FocusEnd: req.GetFocusEnd(), Limit: req.GetLimit(),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AlertAnalysisFingerprintItem, 0, len(fingerprints.Items))
	for _, item := range fingerprints.Items {
		items = append(items, &v1.AlertAnalysisFingerprintItem{
			Fingerprint: item.Fingerprint, RuleUid: item.RuleUID, RuleName: item.RuleName, Events: item.Events,
		})
	}
	return &v1.GetAlertAnalysisFingerprintsResponse{Complete: fingerprints.Complete, Items: items}, nil
}

func (s *AlertEventService) GetAlertAnalysisRecovery(ctx context.Context, req *v1.GetAlertAnalysisRecoveryRequest) (*v1.GetAlertAnalysisRecoveryResponse, error) {
	recovery, err := s.coordinator.AnalysisRecovery(ctx, alertevent.AnalysisQuery{
		From: req.GetFrom(), To: req.GetTo(), TimeZone: req.GetTimeZone(), Status: req.GetStatus(),
		FocusStart: req.GetFocusStart(), FocusEnd: req.GetFocusEnd(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertAnalysisRecoveryResponse{
		Complete: recovery.Complete,
		Auto:     alertAnalysisRecoveryStats(recovery.Auto),
		Manual:   alertAnalysisRecoveryStats(recovery.Manual),
	}, nil
}

// alertAnalysisRecoveryStats 把闭环时长统计转换成接口结构。
func alertAnalysisRecoveryStats(stats bo.AlertAnalysisRecoveryStats) *v1.AlertAnalysisRecoveryStats {
	return &v1.AlertAnalysisRecoveryStats{
		MedianSeconds: stats.MedianSeconds,
		P90Seconds:    stats.P90Seconds,
		Samples:       stats.Samples,
	}
}

func parseAlertEventTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, merr.Invalid("invalid alert event time")
	}
	return &parsed, nil
}

func (s *AlertEventService) GetAlertEvent(ctx context.Context, req *v1.GetAlertEventRequest) (*v1.GetAlertEventResponse, error) {
	item, err := s.coordinator.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertEventResponse{Event: eventItem(*item)}, nil
}

func (s *AlertEventService) GetAlertEventChart(ctx context.Context, req *v1.GetAlertEventChartRequest) (*v1.GetAlertEventChartResponse, error) {
	chart, err := s.coordinator.Chart(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertEventChartResponse{
		DatasourceName: chart.DatasourceName, Expr: chart.Expr,
		Start: chart.Start, End: chart.End, Step: chart.Step, Body: chart.Body,
	}, nil
}

func (s *AlertEventService) AlertEventStats(ctx context.Context, _ *v1.AlertEventStatsRequest) (*v1.AlertEventStatsResponse, error) {
	stats, err := s.coordinator.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.AlertEventStatsResponse{Firing: int32(stats.Firing), Suppressed: int32(stats.Suppressed), Recovered: int32(stats.Recovered)}, nil
}

func (s *AlertEventService) AlertEventRealtimeStats(ctx context.Context, _ *v1.AlertEventRealtimeStatsRequest) (*v1.AlertEventRealtimeStatsResponse, error) {
	stats, err := s.coordinator.RealtimeStats(ctx)
	if err != nil {
		return nil, err
	}
	tabs := make([]*v1.AlertEventRealtimeTabStat, 0, len(stats.Tabs))
	for _, tab := range stats.Tabs {
		tabs = append(tabs, &v1.AlertEventRealtimeTabStat{Id: tab.ID, Count: int32(tab.Count)})
	}
	return &v1.AlertEventRealtimeStatsResponse{
		Tabs:       tabs,
		Firing:     int32(stats.Firing),
		Suppressed: int32(stats.Suppressed),
		Recovered:  int32(stats.Recovered),
	}, nil
}

func (s *AlertEventService) InterveneAlertEvent(ctx context.Context, req *v1.InterveneAlertEventRequest) (*v1.InterveneAlertEventResponse, error) {
	if err := s.coordinator.Intervene(ctx, []string{req.GetUid()}); err != nil {
		return nil, err
	}
	return &v1.InterveneAlertEventResponse{}, nil
}

func (s *AlertEventService) InterveneAlertEvents(ctx context.Context, req *v1.InterveneAlertEventsRequest) (*v1.InterveneAlertEventsResponse, error) {
	if err := s.coordinator.Intervene(ctx, req.GetUids()); err != nil {
		return nil, err
	}
	return &v1.InterveneAlertEventsResponse{}, nil
}

func (s *AlertEventService) SuppressAlertEvent(ctx context.Context, req *v1.SuppressAlertEventRequest) (*v1.SuppressAlertEventResponse, error) {
	until, err := s.coordinator.Suppress(ctx, req.GetUid(), req.GetSuppressedSeconds(), req.GetPermanent())
	if err != nil {
		return nil, err
	}
	if until.IsZero() {
		// 永久抑制没有到期时间，返回空字符串让前端显示「永久」。
		return &v1.SuppressAlertEventResponse{}, nil
	}
	return &v1.SuppressAlertEventResponse{SuppressedUntil: until.UTC().Format(time.RFC3339Nano)}, nil
}

func (s *AlertEventService) RecoverAlertEvent(ctx context.Context, req *v1.RecoverAlertEventRequest) (*v1.RecoverAlertEventResponse, error) {
	if err := s.coordinator.Recover(ctx, []string{req.GetUid()}); err != nil {
		return nil, err
	}
	return &v1.RecoverAlertEventResponse{}, nil
}

func (s *AlertEventService) RecoverAlertEvents(ctx context.Context, req *v1.RecoverAlertEventsRequest) (*v1.RecoverAlertEventsResponse, error) {
	if err := s.coordinator.Recover(ctx, req.GetUids()); err != nil {
		return nil, err
	}
	return &v1.RecoverAlertEventsResponse{}, nil
}

func eventItems(items []bo.AlertEvent) []*v1.AlertEventItem {
	out := make([]*v1.AlertEventItem, 0, len(items))
	for _, item := range items {
		out = append(out, eventItem(item))
	}
	return out
}

func eventItem(item bo.AlertEvent) *v1.AlertEventItem {
	ends := ""
	if item.EndsAt != nil {
		ends = item.EndsAt.UTC().Format(time.RFC3339Nano)
	}
	intervenedAt := ""
	if item.IntervenedAt != nil {
		intervenedAt = item.IntervenedAt.UTC().Format(time.RFC3339Nano)
	}
	suppressedUntil := ""
	if item.SuppressedUntil != nil {
		suppressedUntil = item.SuppressedUntil.UTC().Format(time.RFC3339Nano)
	}
	return &v1.AlertEventItem{
		Uid: item.UID, RuleUid: item.RuleUID, RuleName: item.RuleName, RuleGroupUid: item.RuleGroupUID,
		DatasourceUid: item.DatasourceUID, DatasourceName: item.DatasourceName, Fingerprint: item.Fingerprint,
		Status: item.Status, Expr: item.Expr, Labels: item.Labels, Summary: item.Summary, Description: item.Description,
		SampleValue: item.SampleValue, LevelUid: item.LevelUID, LevelName: item.LevelName, StartsAt: item.StartsAt.UTC().Format(time.RFC3339Nano),
		EndsAt: ends, EvalSlot: item.EvalSlot.UTC().Format(time.RFC3339Nano), Suppressed: item.Suppressed,
		SuppressedUntil: suppressedUntil,
		IntervenedBy:    item.IntervenedBy, IntervenedName: item.IntervenedName, IntervenedAt: intervenedAt,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func parseSlot(value string) (time.Time, error) {
	slot, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, merr.Invalid("invalid request")
	}
	return slot.UTC(), nil
}

// parseOptionalSlot 解析可以为空的时间字段，空串表示这个时间还没有。
func parseOptionalSlot(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	slot, err := parseSlot(value)
	if err != nil {
		return nil, err
	}
	return &slot, nil
}
