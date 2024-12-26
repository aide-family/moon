package service

import (
	"context"

	"github.com/aide-family/moon/api"
	alertapi "github.com/aide-family/moon/api/houyi/alert"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"
)

// AlertService 告警服务
type AlertService struct {
	alertapi.UnimplementedPushAlertServer
	api.UnimplementedAlertServer

	alertBiz    *biz.AlertBiz
	strategyBiz *biz.StrategyBiz
}

// NewAlertService 创建告警服务
func NewAlertService(alertBiz *biz.AlertBiz, strategyBiz *biz.StrategyBiz) *AlertService {
	return &AlertService{
		alertBiz:    alertBiz,
		strategyBiz: strategyBiz,
	}
}

// Hook 告警hook
func (s *AlertService) Hook(ctx context.Context, req *api.AlarmItem) (*api.HookReply, error) {
	if len(req.GetAlerts()) == 0 {
		return &api.HookReply{}, nil
	}
	alarmBo := build.NewAlarmAPIBuilder(req).ToBo()
	if err := s.alertBiz.PushAlarm(ctx, alarmBo); err != nil {
		return nil, err
	}
	return &api.HookReply{}, nil
}

// Alarm 告警
func (s *AlertService) Alarm(ctx context.Context, req *alertapi.AlarmRequest) (*alertapi.AlarmReply, error) {
	strategyInfo := build.NewMetricStrategyBuilder(req.GetStrategy()).ToBo()
	innerAlarm, err := s.InnerAlarm(ctx, strategyInfo)
	if err != nil {
		return nil, err
	}
	if err := s.alertBiz.SaveAlarm(ctx, innerAlarm); err != nil {
		return nil, err
	}
	alarm := build.NewAlarmBuilder(innerAlarm).ToAPI()
	return &alertapi.AlarmReply{
		Alarm: alarm,
	}, nil
}

// InnerAlarm 内部告警
func (s *AlertService) InnerAlarm(ctx context.Context, req bo.IStrategy) (*bo.Alarm, error) {
	return s.strategyBiz.Eval(ctx, req)
}
