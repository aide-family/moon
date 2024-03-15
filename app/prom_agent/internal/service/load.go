package service

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
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
	groups := generateGroups(req.GetGroupList())
	alerting := strategy.NewAlerting(groups...)
	alarms, err := alerting.Eval(context.Background())
	if err != nil {
		s.log.Error("eval error ", err)
		return nil, err
	}

	if len(alarms) == 0 {
		s.log.Debugf("eval success, but no alarm")
		return &pb.EvaluateReply{}, nil
	}

	if err = s.alarmBiz.SendAlarm(alarms...); err != nil {
		s.log.Error("send alarm error ", err)
		return nil, err
	}

	return &pb.EvaluateReply{}, nil
}

// generateGroups 处理groupList
func generateGroups(groupList []*api.GroupSimple) []*strategy.Group {
	groups := make([]*strategy.Group, 0, len(groupList))
	for _, group := range groupList {
		groupInfo := &strategy.Group{
			Name: group.GetName(),
			Id:   group.GetId(),
		}
		for _, strategyInfo := range group.GetStrategies() {
			duration := strategyInfo.GetDuration()
			ruleLabels := strategyInfo.GetLabels()
			if ruleLabels == nil {
				ruleLabels = make(map[string]string)
			}
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
		groups = append(groups, groupInfo)
	}
	return groups
}
