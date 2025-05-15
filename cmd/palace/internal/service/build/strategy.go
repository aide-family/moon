package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ToSaveTeamStrategyParams(request *palace.SaveTeamStrategyRequest) *bo.SaveTeamStrategyParams {
	if validate.IsNil(request) {
		panic("SaveTeamStrategyRequest is nil")
	}
	return &bo.SaveTeamStrategyParams{
		ID:              request.GetStrategyId(),
		Name:            request.GetName(),
		Remark:          request.GetRemark(),
		StrategyType:    vobj.StrategyType(request.GetStrategyType()),
		ReceiverRoutes:  request.GetReceiverRoutes(),
		StrategyGroupID: request.GetGroupId(),
	}
}

func ToSaveTeamMetricStrategyParams(request *palace.SaveTeamMetricStrategyRequest) *bo.SaveTeamMetricStrategyParams {
	if validate.IsNil(request) {
		panic("SaveTeamMetricStrategyRequest is nil")
	}
	return &bo.SaveTeamMetricStrategyParams{
		StrategyID:  request.GetStrategyId(),
		ID:          request.GetMetricStrategyId(),
		Expr:        request.GetExpr(),
		Labels:      request.GetLabels(),
		Annotations: request.GetAnnotations(),
		Datasource:  request.GetDatasource(),
	}
}

func ToSaveTeamMetricStrategyLevelsParams(request *palace.SaveTeamMetricStrategyLevelsRequest) *bo.SaveTeamMetricStrategyLevelsParams {
	if validate.IsNil(request) {
		panic("SaveTeamMetricStrategyLevelsRequest is nil")
	}
	return &bo.SaveTeamMetricStrategyLevelsParams{
		StrategyMetricID: request.GetStrategyMetricId(),
		Levels:           slices.Map(request.GetLevels(), ToSaveTeamMetricStrategyLevelParams),
	}
}

func ToSaveTeamMetricStrategyLevelParams(request *palace.SaveTeamMetricStrategyLevelRequest) *bo.SaveTeamMetricStrategyLevelParams {
	if validate.IsNil(request) {
		panic("SaveTeamMetricStrategyLevelRequest is nil")
	}
	return &bo.SaveTeamMetricStrategyLevelParams{
		ID:             request.GetStrategyMetricLevelId(),
		LevelId:        request.GetLevelId(),
		LevelName:      request.GetLevelName(),
		SampleMode:     vobj.SampleMode(request.GetSampleMode()),
		Total:          request.GetTotal(),
		Condition:      vobj.ConditionMetric(request.GetCondition()),
		Values:         request.GetValues(),
		ReceiverRoutes: request.GetReceiverRoutes(),
		LabelNotices:   slices.Map(request.GetLabelNotices(), ToLabelNoticeParams),
		Duration:       request.GetDuration().AsDuration(),
		Status:         vobj.GlobalStatus(request.GetStatus()),
		AlarmPages:     request.GetAlarmPages(),
	}
}

func ToLabelNoticeParams(request *palace.LabelNotices) *bo.LabelNoticeParams {
	if validate.IsNil(request) {
		panic("LabelNotices is nil")
	}
	return &bo.LabelNoticeParams{
		Key:            request.GetKey(),
		Value:          request.GetValue(),
		ReceiverRoutes: request.GetReceiverRoutes(),
	}
}

func ToUpdateTeamStrategiesStatusParams(request *palace.UpdateTeamStrategiesStatusRequest) *bo.UpdateTeamStrategiesStatusParams {
	if validate.IsNil(request) {
		panic("UpdateTeamStrategiesStatusRequest is nil")
	}
	return &bo.UpdateTeamStrategiesStatusParams{
		StrategyIds: request.GetStrategyIds(),
		Status:      vobj.GlobalStatus(request.GetStatus()),
	}
}

func ToOperateTeamStrategyParams(request *palace.OperateTeamStrategyRequest) *bo.OperateTeamStrategyParams {
	if validate.IsNil(request) {
		panic("OperateTeamStrategyRequest is nil")
	}
	return &bo.OperateTeamStrategyParams{
		StrategyId: request.GetStrategyId(),
	}
}

func ToListTeamStrategyParams(request *palace.ListTeamStrategyRequest) *bo.ListTeamStrategyParams {
	if validate.IsNil(request) {
		panic("ListTeamStrategyRequest is nil")
	}
	return &bo.ListTeamStrategyParams{
		PaginationRequest: ToPaginationRequest(request.GetPagination()),
		Keyword:           request.GetKeyword(),
		Status:            vobj.GlobalStatus(request.GetStatus()),
		GroupIds:          request.GetGroupIds(),
		StrategyTypes:     slices.Map(request.GetStrategyTypes(), func(strategyType common.StrategyType) vobj.StrategyType { return vobj.StrategyType(strategyType) }),
	}
}

func ToSubscribeTeamStrategyParams(request *palace.SubscribeTeamStrategyRequest) *bo.SubscribeTeamStrategyParams {
	if validate.IsNil(request) {
		panic("SubscribeTeamStrategyRequest is nil")
	}
	return &bo.SubscribeTeamStrategyParams{
		StrategyId: request.GetStrategyId(),
		NoticeType: vobj.NoticeType(request.GetSubscribeType()),
	}
}

func ToSubscribeTeamStrategiesParams(request *palace.SubscribeTeamStrategiesRequest) *bo.SubscribeTeamStrategiesParams {
	if validate.IsNil(request) {
		panic("SubscribeTeamStrategiesRequest is nil")
	}
	return &bo.SubscribeTeamStrategiesParams{
		StrategyId:        request.GetStrategyId(),
		PaginationRequest: ToPaginationRequest(request.GetPagination()),
		Subscribers:       request.GetSubscribers(),
		NoticeType:        vobj.NoticeType(request.GetSubscribeType()),
	}
}

func ToTeamStrategyItem(strategy do.Strategy) *common.TeamStrategyItem {
	if validate.IsNil(strategy) {
		return nil
	}
	return &common.TeamStrategyItem{
		StrategyId:   strategy.GetID(),
		GroupId:      strategy.GetStrategyGroupID(),
		Name:         strategy.GetName(),
		Remark:       strategy.GetRemark(),
		Status:       common.GlobalStatus(strategy.GetStatus()),
		Creator:      ToUserBaseItem(strategy.GetCreator()),
		CreatedAt:    timex.Format(strategy.GetCreatedAt()),
		UpdatedAt:    timex.Format(strategy.GetUpdatedAt()),
		Team:         ToTeamBaseItem(strategy.GetTeam()),
		Notices:      ToNoticeGroupItems(strategy.GetNotices()),
		StrategyType: common.StrategyType(strategy.GetStrategyType().GetValue()),
		Group:        ToTeamStrategyGroupItem(strategy.GetStrategyGroup()),
	}
}

func ToTeamStrategyItems(strategies []do.Strategy) []*common.TeamStrategyItem {
	return slices.Map(strategies, ToTeamStrategyItem)
}

func ToTeamMetricStrategyItem(strategy do.StrategyMetric) *common.TeamStrategyMetricItem {
	if validate.IsNil(strategy) {
		panic("do.StrategyMetric is nil")
	}
	return &common.TeamStrategyMetricItem{
		Base:                ToTeamStrategyItem(strategy.GetStrategy()),
		StrategyMetricId:    strategy.GetID(),
		Expr:                strategy.GetExpr(),
		Labels:              strategy.GetLabels(),
		Annotations:         strategy.GetAnnotations(),
		StrategyMetricRules: ToTeamMetricStrategyItemRules(strategy.GetRules()),
		Datasource:          ToTeamMetricDatasourceItems(strategy.GetDatasourceList()),
		Creator:             ToUserBaseItem(strategy.GetCreator()),
	}
}

func ToTeamMetricStrategyItems(strategies []do.StrategyMetric) []*common.TeamStrategyMetricItem {
	return slices.Map(strategies, ToTeamMetricStrategyItem)
}

func ToTeamMetricStrategyItemRule(rule do.StrategyMetricRule) *common.TeamStrategyMetricItem_RuleItem {
	if validate.IsNil(rule) {
		return nil
	}
	return &common.TeamStrategyMetricItem_RuleItem{
		RuleId:           rule.GetID(),
		StrategyMetricId: rule.GetStrategyMetricID(),
		Level:            ToDictItem(rule.GetLevel()),
		SampleMode:       common.SampleMode(rule.GetSampleMode()),
		Condition:        common.ConditionMetric(rule.GetCondition()),
		Total:            rule.GetTotal(),
		Values:           rule.GetValues(),
		Duration:         durationpb.New(rule.GetDuration()),
		Status:           common.GlobalStatus(rule.GetStatus().GetValue()),
		Notices:          ToNoticeGroupItems(rule.GetNotices()),
		LabelNotices:     ToLabelNoticeItems(rule.GetLabelNotices()),
		AlarmPages:       ToDictItems(rule.GetAlarmPages()),
	}
}

func ToTeamMetricStrategyItemRules(rules []do.StrategyMetricRule) []*common.TeamStrategyMetricItem_RuleItem {
	return slices.Map(rules, ToTeamMetricStrategyItemRule)
}

func ToLabelNoticeItems(labelNotices []do.StrategyMetricRuleLabelNotice) []*common.StrategyMetricRuleLabelNotice {
	return slices.Map(labelNotices, ToLabelNoticeItem)
}

func ToLabelNoticeItem(labelNotice do.StrategyMetricRuleLabelNotice) *common.StrategyMetricRuleLabelNotice {
	if validate.IsNil(labelNotice) {
		return nil
	}
	return &common.StrategyMetricRuleLabelNotice{
		LabelNoticeId:        labelNotice.GetID(),
		CreatedAt:            timex.Format(labelNotice.GetCreatedAt()),
		UpdatedAt:            timex.Format(labelNotice.GetUpdatedAt()),
		StrategyMetricRuleId: labelNotice.GetStrategyMetricRuleID(),
		LabelKey:             labelNotice.GetLabelKey(),
		LabelValue:           labelNotice.GetLabelValue(),
		Notices:              ToNoticeGroupItems(labelNotice.GetNotices()),
	}
}

func ToSubscribeTeamStrategiesItems(subscribers []do.TeamStrategySubscriber) []*common.SubscriberItem {
	return slices.Map(subscribers, ToSubscribeTeamStrategyItem)
}

func ToSubscribeTeamStrategyItem(subscriber do.TeamStrategySubscriber) *common.SubscriberItem {
	if validate.IsNil(subscriber) {
		return nil
	}
	return &common.SubscriberItem{
		User:          ToUserBaseItem(subscriber.GetCreator()),
		SubscribeType: common.NoticeType(subscriber.GetSubscribeType().GetValue()),
		Strategy:      ToTeamStrategyItem(subscriber.GetStrategy()),
		SubscribeTime: timex.Format(subscriber.GetCreatedAt()),
	}
}
