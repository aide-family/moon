package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aide-family/moon/api"
	pb "github.com/aide-family/moon/api/agent"
	"github.com/aide-family/moon/app/prom_agent/internal/biz"
	"github.com/aide-family/moon/app/prom_agent/internal/biz/bo"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/go-kratos/kratos/v2/log"
)

type LoadService struct {
	pb.UnimplementedLoadServer

	log         *log.Helper
	alarmBiz    *biz.AlarmBiz
	evaluateBiz *biz.EvaluateBiz
}

func NewLoadService(alarmBiz *biz.AlarmBiz, evaluateBiz *biz.EvaluateBiz, logger log.Logger) *LoadService {
	return &LoadService{
		log:         log.NewHelper(log.With(logger, "module", "service.load")),
		alarmBiz:    alarmBiz,
		evaluateBiz: evaluateBiz,
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
		//s.log.Debugf("eval success, but no alarm")
		return &pb.EvaluateReply{}, nil
	}

	if err = s.alarmBiz.SendAlarm(alarms...); err != nil {
		s.log.Error("send alarm error ", err)
		return nil, err
	}

	return &pb.EvaluateReply{}, nil
}

// EvaluateV2 批量处理groupList
func (s *LoadService) EvaluateV2(ctx context.Context, req *pb.EvaluateV2Request) (*pb.EvaluateV2Reply, error) {
	alarmList, err := s.evaluateBiz.EvaluateV2(ctx, evaluateV2RequestToBo(req.GetGroupList()))
	if err != nil {
		s.log.Errorw("evaluateV2 error err", err)
		return nil, err
	}

	if len(alarmList) > 0 {
		if err = s.alarmBiz.SendAlarmV2(ctx, alarmList...); err != nil {
			s.log.Errorw("send alarmV2 error err", err)
			return nil, err
		}
	}

	return &pb.EvaluateV2Reply{
		Message: "success",
		Code:    0,
	}, nil
}

// evaluateV2RequestToBo 处理groupList
func evaluateV2RequestToBo(groupList []*api.EvaluateGroup) *bo.EvaluateReqBo {
	groups := make([]*bo.EvaluateStrategyGroup, 0, len(groupList))
	for _, groupItem := range groupList {
		item := &bo.EvaluateStrategyGroup{
			GroupId:      groupItem.GetId(),
			GroupName:    groupItem.GetName(),
			StrategyList: make([]*bo.EvaluateStrategy, 0, len(groupItem.GetStrategies())),
		}

		for _, strategyItem := range groupItem.GetStrategies() {
			datasourceInfo := strategyItem.GetDatasource()
			item.StrategyList = append(item.StrategyList, &bo.EvaluateStrategy{
				Id:          strategyItem.GetId(),
				Alert:       strategyItem.GetAlert(),
				Expr:        strategyItem.GetExpr(),
				For:         fmt.Sprintf("%d%s", strategyItem.GetDuration().GetValue(), strategyItem.GetDuration().GetUnit()),
				Labels:      strategyItem.GetLabels(),
				Annotations: strategyItem.GetAnnotations(),
				Datasource: datasource.NewDataSource(
					datasource.WithPrometheusConfig(
						p8s.WithBasicAuth(agent.NewBasicAuthWithString(datasourceInfo.GetBasicAuth())),
						p8s.WithEndpoint(datasourceInfo.GetEndpoint()),
					),
				),
			})
		}
		groups = append(groups, item)
	}

	return &bo.EvaluateReqBo{
		GroupList: groups,
	}
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
			ruleInfo.SetBasicAuth(strategy.NewBasicAuthWithString(strategyInfo.GetBasicAuth()))
			groupInfo.Rules = append(groupInfo.Rules, ruleInfo)
		}
		groups = append(groups, groupInfo)
	}
	return groups
}
