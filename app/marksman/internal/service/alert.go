package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

func NewAlertService(
	alertPageBiz *biz.AlertPageBiz,
	alertBiz *biz.AlertBiz,
) *AlertService {
	return &AlertService{
		alertPageBiz: alertPageBiz,
		alertBiz:     alertBiz,
	}
}

type AlertService struct {
	apiv1.UnimplementedAlertServer

	alertPageBiz *biz.AlertPageBiz
	alertBiz     *biz.AlertBiz
}

func (s *AlertService) CreateAlertPage(ctx context.Context, req *apiv1.CreateAlertPageRequest) (*apiv1.CreateAlertPageReply, error) {
	uid, err := s.alertPageBiz.CreateAlertPage(ctx, bo.NewCreateAlertPageBo(req))
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateAlertPageReply{Uid: uid.Int64()}, nil
}

func (s *AlertService) UpdateAlertPage(ctx context.Context, req *apiv1.UpdateAlertPageRequest) (*apiv1.UpdateAlertPageReply, error) {
	if err := s.alertPageBiz.UpdateAlertPage(ctx, bo.NewUpdateAlertPageBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateAlertPageReply{}, nil
}

func (s *AlertService) DeleteAlertPage(ctx context.Context, req *apiv1.DeleteAlertPageRequest) (*apiv1.DeleteAlertPageReply, error) {
	if err := s.alertPageBiz.DeleteAlertPage(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteAlertPageReply{}, nil
}

func (s *AlertService) GetAlertPage(ctx context.Context, req *apiv1.GetAlertPageRequest) (*apiv1.AlertPageItem, error) {
	item, err := s.alertPageBiz.GetAlertPage(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1AlertPageItem(item), nil
}

func (s *AlertService) ListAlertPage(ctx context.Context, req *apiv1.ListAlertPageRequest) (*apiv1.ListAlertPageReply, error) {
	result, err := s.alertPageBiz.ListAlertPage(ctx, bo.NewListAlertPageBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListAlertPageReply(result), nil
}

func (s *AlertService) ListRealtimeAlert(ctx context.Context, req *apiv1.ListRealtimeAlertRequest) (*apiv1.ListRealtimeAlertReply, error) {
	result, err := s.alertBiz.ListRealtimeAlert(ctx, bo.NewListRealtimeAlertBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListRealtimeAlertReply(result), nil
}

func (s *AlertService) ListHistoryAlert(ctx context.Context, req *apiv1.ListHistoryAlertRequest) (*apiv1.ListHistoryAlertReply, error) {
	reqBo := bo.NewListHistoryAlertBo(req)
	result, err := s.alertBiz.ListRealtimeAlert(ctx, reqBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListHistoryAlertReply(result), nil
}

func (s *AlertService) GetAlertEvent(ctx context.Context, req *apiv1.GetAlertEventRequest) (*apiv1.AlertEventItem, error) {
	item, err := s.alertBiz.GetAlertEvent(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1AlertEventItem(item), nil
}

func (s *AlertService) InterveneAlert(ctx context.Context, req *apiv1.InterveneAlertRequest) (*apiv1.InterveneAlertReply, error) {
	reqBo := bo.NewInterveneAlertBo(ctx, req)
	if err := s.alertBiz.InterveneAlert(ctx, reqBo); err != nil {
		return nil, err
	}
	return &apiv1.InterveneAlertReply{}, nil
}

func (s *AlertService) BatchInterveneAlert(ctx context.Context, req *apiv1.BatchInterveneAlertRequest) (*apiv1.BatchInterveneAlertReply, error) {
	reqBo := bo.NewBatchInterveneAlertBo(ctx, req)
	if err := s.alertBiz.BatchInterveneAlert(ctx, reqBo); err != nil {
		return nil, err
	}
	return &apiv1.BatchInterveneAlertReply{}, nil
}

func (s *AlertService) SuppressAlert(ctx context.Context, req *apiv1.SuppressAlertRequest) (*apiv1.SuppressAlertReply, error) {
	reqBo := bo.NewSuppressAlertBo(ctx, req)
	if err := s.alertBiz.SuppressAlert(ctx, reqBo); err != nil {
		return nil, err
	}
	return &apiv1.SuppressAlertReply{}, nil
}

func (s *AlertService) RecoverAlert(ctx context.Context, req *apiv1.RecoverAlertRequest) (*apiv1.RecoverAlertReply, error) {
	reqBo := bo.NewRecoverAlertBo(ctx, req)
	if err := s.alertBiz.RecoverAlert(ctx, reqBo); err != nil {
		return nil, err
	}
	return &apiv1.RecoverAlertReply{}, nil
}

func (s *AlertService) BatchRecoverAlert(ctx context.Context, req *apiv1.BatchRecoverAlertRequest) (*apiv1.BatchRecoverAlertReply, error) {
	reqBo := bo.NewBatchRecoverAlertBo(ctx, req)
	if err := s.alertBiz.BatchRecoverAlert(ctx, reqBo); err != nil {
		return nil, err
	}
	return &apiv1.BatchRecoverAlertReply{}, nil
}

func (s *AlertService) GetAlertStatistics(ctx context.Context, req *apiv1.GetAlertStatisticsRequest) (*apiv1.GetAlertStatisticsReply, error) {
	stats, err := s.alertBiz.GetAlertStatistics(ctx)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1GetAlertStatisticsReply(stats), nil
}

func (s *AlertService) ListUserAlertPages(ctx context.Context, req *apiv1.ListUserAlertPagesRequest) (*apiv1.ListUserAlertPagesReply, error) {
	items, err := s.alertPageBiz.ListUserAlertPages(ctx)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListUserAlertPagesReply(items), nil
}

func (s *AlertService) SaveUserAlertPages(ctx context.Context, req *apiv1.SaveUserAlertPagesRequest) (*apiv1.SaveUserAlertPagesReply, error) {
	uids := make([]snowflake.ID, 0, len(req.GetAlertPageUids()))
	for _, id := range req.GetAlertPageUids() {
		uids = append(uids, snowflake.ParseInt64(id))
	}
	if err := s.alertPageBiz.SaveUserAlertPages(ctx, uids); err != nil {
		return nil, err
	}
	return &apiv1.SaveUserAlertPagesReply{}, nil
}
