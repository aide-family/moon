package service

import (
	"context"

	alertapi "github.com/aide-family/moon/api/houyi/alert"
)

type AlertService struct {
	alertapi.UnimplementedAlertServer
}

func NewAlertService() *AlertService {
	return &AlertService{}
}

func (s *AlertService) Hook(ctx context.Context, req *alertapi.HookRequest) (*alertapi.HookReply, error) {
	return &alertapi.HookReply{}, nil
}
