package service

import (
	"context"

	"github.com/aide-family/moon/api"
	alertapi "github.com/aide-family/moon/api/houyi/alert"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"
	"github.com/go-kratos/kratos/v2/log"
)

type AlertService struct {
	alertapi.UnimplementedAlertServer

	alertBiz    *biz.AlertBiz
	strategyBiz *biz.StrategyBiz
}

func NewAlertService(alertBiz *biz.AlertBiz, strategyBiz *biz.StrategyBiz) *AlertService {
	return &AlertService{
		alertBiz:    alertBiz,
		strategyBiz: strategyBiz,
	}
}

func (s *AlertService) Hook(ctx context.Context, req *api.AlarmItem) (*alertapi.HookReply, error) {
	return &alertapi.HookReply{}, nil
}

// Alarm 告警
func (s *AlertService) Alarm(ctx context.Context, req *alertapi.AlarmRequest) (*alertapi.AlarmReply, error) {
	strategyInfo := build.NewStrategyBuilder(req.GetStrategy()).ToBo()
	innerAlarm, err := s.InnerAlarm(ctx, strategyInfo)
	if err != nil {
		return nil, err
	}
	if err := s.alertBiz.SaveAlarm(ctx, innerAlarm); err != nil {
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
	return s.strategyBiz.Eval(ctx, req)
}
