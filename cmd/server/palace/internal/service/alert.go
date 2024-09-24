package service

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// AlertService alert service
type AlertService struct {
	api.UnimplementedAlertServer

	alertBiz    *biz.AlarmBiz
	strategyBiz *biz.StrategyBiz
}

// NewAlertService 创建告警服务
func NewAlertService(alertBiz *biz.AlarmBiz, strategyBiz *biz.StrategyBiz) *AlertService {
	return &AlertService{
		alertBiz:    alertBiz,
		strategyBiz: strategyBiz,
	}
}

// InnerAlarm 内部告警
func (s *AlertService) InnerAlarm(ctx context.Context, req *bo.Strategy) (*bo.Alarm, error) {
	log.Debugw("InnerAlarm", req)
	return s.strategyBiz.Eval(ctx, req)
}

// PushStrategy 推送策略
func (s *AlertService) PushStrategy(ctx context.Context, strategies []*bo.Strategy) error {
	return s.strategyBiz.PushStrategy(ctx, strategies)
}

// Hook 告警hook
func (s *AlertService) Hook(_ context.Context, req *api.AlarmItem) (*api.HookReply, error) {
	param := builder.NewParamsBuild().
		AlarmModuleBuilder().
		WithCreateAlarmRawInfoRequest(req).
		ToBo()
	err := s.alertBiz.SaveAlertQueue(param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &api.HookReply{
		Msg:  "success",
		Code: 200,
	}, nil
}

// CreateAlarmInfo 创建告警信息
func (s *AlertService) CreateAlarmInfo(ctx context.Context, params *bo.CreateAlarmHookRawParams) error {
	return s.alertBiz.CreateAlarmInfo(ctx, params)
}
