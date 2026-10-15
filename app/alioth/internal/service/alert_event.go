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
	events *alertevent.Usecase
}

func NewAlertEventService(uc *alertevent.Usecase) *AlertEventService {
	return &AlertEventService{events: uc}
}

func (s *AlertEventService) ListAlertEvents(ctx context.Context, req *v1.ListAlertEventsRequest) (*v1.ListAlertEventsResponse, error) {
	items, total, page, pageSize, err := s.events.List(ctx, req.GetView(), req.GetName(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertEventsResponse{Items: eventItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *AlertEventService) GetAlertEvent(ctx context.Context, req *v1.GetAlertEventRequest) (*v1.GetAlertEventResponse, error) {
	item, err := s.events.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertEventResponse{Event: eventItem(*item)}, nil
}

func (s *AlertEventService) AlertEventStats(ctx context.Context, _ *v1.AlertEventStatsRequest) (*v1.AlertEventStatsResponse, error) {
	stats, err := s.events.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.AlertEventStatsResponse{Firing: int32(stats.Firing), Suppressed: int32(stats.Suppressed), Recovered: int32(stats.Recovered)}, nil
}

func (s *AlertEventService) InterveneAlertEvent(ctx context.Context, req *v1.InterveneAlertEventRequest) (*v1.InterveneAlertEventResponse, error) {
	if err := s.events.Intervene(ctx, []string{req.GetUid()}); err != nil {
		return nil, err
	}
	return &v1.InterveneAlertEventResponse{}, nil
}

func (s *AlertEventService) InterveneAlertEvents(ctx context.Context, req *v1.InterveneAlertEventsRequest) (*v1.InterveneAlertEventsResponse, error) {
	if err := s.events.Intervene(ctx, req.GetUids()); err != nil {
		return nil, err
	}
	return &v1.InterveneAlertEventsResponse{}, nil
}

func (s *AlertEventService) SuppressAlertEvent(ctx context.Context, req *v1.SuppressAlertEventRequest) (*v1.SuppressAlertEventResponse, error) {
	if err := s.events.Suppress(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.SuppressAlertEventResponse{}, nil
}

func (s *AlertEventService) RecoverAlertEvent(ctx context.Context, req *v1.RecoverAlertEventRequest) (*v1.RecoverAlertEventResponse, error) {
	if err := s.events.Recover(ctx, []string{req.GetUid()}); err != nil {
		return nil, err
	}
	return &v1.RecoverAlertEventResponse{}, nil
}

func (s *AlertEventService) RecoverAlertEvents(ctx context.Context, req *v1.RecoverAlertEventsRequest) (*v1.RecoverAlertEventsResponse, error) {
	if err := s.events.Recover(ctx, req.GetUids()); err != nil {
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
	return &v1.AlertEventItem{
		Uid: item.UID, RuleUid: item.RuleUID, DatasourceUid: item.DatasourceUID, Fingerprint: item.Fingerprint,
		Status: item.Status, Expr: item.Expr, Labels: item.Labels, Summary: item.Summary, Description: item.Description,
		SampleValue: item.SampleValue, LevelUid: item.LevelUID, StartsAt: item.StartsAt.UTC().Format(time.RFC3339Nano),
		EndsAt: ends, EvalSlot: item.EvalSlot.UTC().Format(time.RFC3339Nano), Suppressed: item.Suppressed,
		IntervenedBy: item.IntervenedBy, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func parseSlot(value string) (time.Time, error) {
	slot, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, merr.Invalid("invalid request")
	}
	return slot.UTC(), nil
}
