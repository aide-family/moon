package service

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz"
	"github.com/aide-family/moon/cmd/rabbit/internal/service/build"
	common "github.com/aide-family/moon/pkg/api/common"
	apicommon "github.com/aide-family/moon/pkg/api/rabbit/common"
	apiv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
)

type AlertService struct {
	apiv1.UnimplementedAlertServer

	alertBiz *biz.Alert
}

func NewAlertService(alertBiz *biz.Alert) *AlertService {
	return &AlertService{
		alertBiz: alertBiz,
	}
}

func (s *AlertService) SendAlert(ctx context.Context, req *common.AlertsItem) (*apicommon.EmptyReply, error) {
	alerts := build.ToAlerts(req)
	if err := s.alertBiz.SendAlert(ctx, alerts); err != nil {
		return nil, err
	}
	return &apicommon.EmptyReply{}, nil
}
