package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	apicommon "github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/merr"
)

func NewAlertService(realtimeBiz *biz.Realtime) *AlertService {
	return &AlertService{
		realtimeBiz: realtimeBiz,
	}
}

type AlertService struct {
	palace.UnimplementedAlertServer
	realtimeBiz *biz.Realtime
}

func (s *AlertService) PushAlert(ctx context.Context, req *apicommon.AlertItem) (*common.EmptyReply, error) {
	params := build.ToAlertParams(req)
	if err := s.realtimeBiz.SaveAlert(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *AlertService) ListAlerts(ctx context.Context, req *palace.ListAlertParams) (*palace.ListAlertReply, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}
	params, err := build.ToListAlertParams(req)
	if err != nil {
		return nil, err
	}
	params.TeamID = teamId
	reply, err := s.realtimeBiz.ListAlerts(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListAlertReply{
		Pagination: build.ToPaginationReply(reply.PaginationReply),
		Items:      build.ToRealtimeAlertItems(reply.Items),
	}, nil
}
