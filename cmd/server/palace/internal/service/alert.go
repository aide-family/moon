package service

import (
	"context"

	"github.com/aide-family/moon/api"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
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
func (s *AlertService) PushStrategy(ctx context.Context, strategies *bo.Strategy) error {
	strategyDetail := builder.NewParamsBuild(ctx).StrategyModuleBuilder().BoStrategyBuilder().ToAPI(strategies)
	return s.strategyBiz.PushStrategy(ctx, strategyDetail)
}

// Hook 告警hook
func (s *AlertService) Hook(ctx context.Context, req *api.AlarmItem) (*api.HookReply, error) {
	param := builder.NewParamsBuild(ctx).AlarmModuleBuilder().WithCreateAlarmRawInfoRequest(req).ToBo()
	err := s.alertBiz.SaveAlertQueue(param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &api.HookReply{Msg: "success", Code: 200}, nil
}

// CreateAlarmInfo 创建告警信息
func (s *AlertService) CreateAlarmInfo(ctx context.Context, params *bo.CreateAlarmHookRawParams) error {
	return s.alertBiz.CreateAlarmInfo(ctx, params)
}

// SendAlertMsg 发送告警消息
func (s *AlertService) SendAlertMsg(ctx context.Context, req *hookapi.SendMsgRequest) error {
	s.alertBiz.SendAlertMsg(ctx, &bo.SendMsg{SendMsgRequest: req})
	return nil
}
