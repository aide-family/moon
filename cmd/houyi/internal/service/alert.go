package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	houyiv1 "github.com/moon-monitor/moon/pkg/api/houyi/v1"
)

type AlertService struct {
	houyiv1.UnimplementedAlertServer

	helper *log.Helper
}

func NewAlertService(logger log.Logger) *AlertService {
	return &AlertService{
		helper: log.NewHelper(log.With(logger, "module", "service.alert")),
	}
}

func (s *AlertService) Push(ctx context.Context, req *houyiv1.PushAlertRequest) (*houyiv1.PushAlertReply, error) {
	return &houyiv1.PushAlertReply{}, nil
}

func (s *AlertService) InnerPush(ctx context.Context, req bo.Alert) {
	bs, err := json.Marshal(req)
	if err == nil {
		s.helper.WithContext(ctx).Debugw("status", req.GetStatus().String(), "alert", string(bs))
	}
}
