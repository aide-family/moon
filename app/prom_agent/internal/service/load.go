package service

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/biz"
	"prometheus-manager/pkg/strategy"
)

type LoadService struct {
	pb.UnimplementedLoadServer

	log      *log.Helper
	alarmBiz *biz.AlarmBiz
}

func NewLoadService(alarmBiz *biz.AlarmBiz, logger log.Logger) *LoadService {
	return &LoadService{
		log:      log.NewHelper(log.With(logger, "module", "service.load")),
		alarmBiz: alarmBiz,
	}
}

func (s *LoadService) Evaluate(_ context.Context, req *pb.EvaluateRequest) (*pb.EvaluateReply, error) {
	alarmList := make([]*strategy.Alarm, 0)
	for _, group := range req.GetGroupList() {
		groupInfo := &strategy.Group{
			Name: group.GetName(),
			Id:   group.GetId(),
		}
		for _, strategyInfo := range group.GetStrategies() {
			duration := strategyInfo.GetDuration()
			ruleLabels := strategyInfo.GetLabels()
			ruleLabels[strategy.MetricLevelId] = strconv.Itoa(int(strategyInfo.GetAlarmLevelId()))
			ruleInfo := &strategy.Rule{
				Id:          strategyInfo.GetId(),
				Alert:       strategyInfo.GetAlert(),
				Expr:        strategyInfo.GetExpr(),
				For:         strconv.Itoa(int(duration.GetValue())) + duration.GetUnit(),
				Labels:      ruleLabels,
				Annotations: strategyInfo.GetAnnotations(),
			}
			ruleInfo.SetEndpoint(strategyInfo.GetEndpoint())
			groupInfo.Rules = append(groupInfo.Rules, ruleInfo)
		}
		alerting := strategy.NewAlerting(groupInfo, strategy.PrometheusDatasource, s.log)
		alarms, err := alerting.Eval(context.Background())
		if err != nil {
			s.log.Error("eval error ", err)
			continue
		}
		alarmList = append(alarmList, alarms...)
	}
	if len(alarmList) > 0 {
		if err := s.alarmBiz.SendAlarm(alarmList...); err != nil {
			s.log.Error("send alarm error ", err)
			return nil, err
		}
	}

	return &pb.EvaluateReply{}, nil
}
