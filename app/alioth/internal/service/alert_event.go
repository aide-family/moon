package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

type AlertEventService struct {
	v1.UnimplementedAlertEventServiceServer
	coordinator *alertevent.Coordinator
}

func NewAlertEventService(coordinator *alertevent.Coordinator) *AlertEventService {
	return &AlertEventService{coordinator: coordinator}
}

func (s *AlertEventService) ListAlertEvents(ctx context.Context, req *v1.ListAlertEventsRequest) (*v1.ListAlertEventsResponse, error) {
	items, total, page, pageSize, err := s.coordinator.List(ctx, req.GetView(), req.GetName(), req.GetMine(), req.GetPageUid(), req.GetLevelUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertEventsResponse{Items: eventItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
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
