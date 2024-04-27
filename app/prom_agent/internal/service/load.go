package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aide-family/moon/api"
	pb "github.com/aide-family/moon/api/agent"
	"github.com/aide-family/moon/app/prom_agent/internal/biz"
	"github.com/aide-family/moon/app/prom_agent/internal/biz/bo"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource"
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

// SendRecoveryEventAlarm 批量处理告警恢复事件
func (s *LoadService) SendRecoveryEventAlarm(ctx context.Context, alarms ...*agent.Alarm) error {
	if pkg.IsNil(alarms) || len(alarms) == 0 {
		return nil
	}

	if err := s.alarmBiz.SendAlarmV2(ctx, alarms...); err != nil {
		s.log.Errorw("send alarmV2 error err", err)
		return err
	}
	return nil
}

// GenerateRecoveryEvent 生成告警恢复事件
func (s *LoadService) GenerateRecoveryEvent(oldGroupInfo, newGroupInfo *api.EvaluateGroup) []*agent.Alarm {
	if oldGroupInfo == nil {
		//  没有旧的， 是新增规则组场景， 不处理
		return nil
	}

	// TODO 告警恢复事件
	// 1. 没有新的， 是删除规则组场景
	var alarmList []*agent.Alarm
	if newGroupInfo == nil {
		for _, strategyItem := range oldGroupInfo.GetStrategies() {
			alarmInfo, err := buildAlarm(strategyItem.GetId())
			if err != nil {
				s.log.Warnw("build alert error", err)
				continue
			}
			alarmList = append(alarmList, alarmInfo)
		}
		return alarmList
	}

	newStrategyMap := make(map[uint32]struct{})
	for _, strategyItem := range newGroupInfo.GetStrategies() {
		newStrategyMap[strategyItem.GetId()] = struct{}{}
	}
	for _, strategyItem := range oldGroupInfo.GetStrategies() {
		if _, ok := newStrategyMap[strategyItem.GetId()]; ok {
			continue
		}
		alarmInfo, err := buildAlarm(strategyItem.GetId())
		if err != nil {
			s.log.Warnw("build alert error", err)
			continue
		}
		alarmList = append(alarmList, alarmInfo)
	}

	return alarmList
}

func buildAlarm(strategyId uint32) (*agent.Alarm, error) {
	cache := agent.GetGlobalCache()
	var alarmInfo agent.Alarm
	if err := cache.Get(strconv.Itoa(int(strategyId)), &alarmInfo); err != nil {
		return nil, err
	}
	alerts := make([]*agent.Alert, 0, len(alarmInfo.Alerts))
	for _, alertInfo := range alarmInfo.Alerts {
		if pkg.IsNil(alertInfo) {
			continue
		}
		alertsTmp := *alertInfo
		alertsTmp.Status = agent.AlarmStatusResolved
		alertsTmp.EndsAt = time.Now().Format(time.RFC3339)
		if err := cache.Delete(alertsTmp.Fingerprint); err != nil {
			return nil, err
		}
		alertsTmp.Fingerprint = alertsTmp.GetMd5Fingerprint()
		alerts = append(alerts, &alertsTmp)
	}

	alarmInfo.Alerts = alerts
	if err := cache.Delete(strconv.Itoa(int(strategyId))); err != nil {
		return nil, err
	}

	return &alarmInfo, cache.Delete(strconv.Itoa(int(strategyId)))
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
		alarms := make([]*agent.Alarm, 0, len(alarmList))
		for _, alarmItem := range alarmList {
			alarmItemTmp := alarmItem
			alerts := make([]*agent.Alert, 0, len(alarmItem.GetAlerts()))
			for _, alertItem := range alarmItem.GetAlerts() {
				alertItemTmp := alertItem
				alertItemTmp.Fingerprint = alertItemTmp.GetMd5Fingerprint()
				alerts = append(alerts, alertItemTmp)
			}
			alarmItemTmp.Alerts = alerts
			alarms = append(alarms, alarmItemTmp)
		}

		if err = s.alarmBiz.SendAlarmV2(ctx, alarms...); err != nil {
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
			datasourceInstance, err := datasource.NewDataSource(
				datasource.WithCategory(agent.DatasourceCategory(datasourceInfo.GetDatasourceType())),
				datasource.WithConfig(&datasource.Config{
					Endpoint:  datasourceInfo.GetEndpoint(),
					BasicAuth: datasourceInfo.GetBasicAuth(),
				}),
			)
			if err != nil {
				log.Warnw("new datasource error", err)
				continue
			}
			item.StrategyList = append(item.StrategyList, &bo.EvaluateStrategy{
				Id:          strategyItem.GetId(),
				Alert:       strategyItem.GetAlert(),
				Expr:        strategyItem.GetExpr(),
				For:         fmt.Sprintf("%d%s", strategyItem.GetDuration().GetValue(), strategyItem.GetDuration().GetUnit()),
				Labels:      strategyItem.GetLabels(),
				Annotations: strategyItem.GetAnnotations(),
				Datasource:  datasourceInstance,
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
