package service

import (
	"context"

	alertapi "github.com/aide-family/moon/api/houyi/alert"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"
	"github.com/go-kratos/kratos/v2/log"
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

// Alarm 告警
func (s *AlertService) Alarm(ctx context.Context, req *alertapi.AlarmRequest) (*alertapi.AlarmReply, error) {
	strategyInfo := build.NewAlarmApiBuilder(req.GetStrategy()).ToBo()
	innerAlarm, err := s.InnerAlarm(ctx, strategyInfo)
	if err != nil {
		return nil, err
	}
	alarm := build.NewAlarmBuilder(innerAlarm).ToApi()
	return &alertapi.AlarmReply{
		Alarm: alarm,
	}, nil
}

// InnerAlarm 内部告警
func (s *AlertService) InnerAlarm(ctx context.Context, req *bo.Strategy) (*bo.Alarm, error) {
	log.Debugw("InnerAlarm", req)
	return &bo.Alarm{}, nil
}
