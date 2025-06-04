package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
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
	labels := slices.Map(request.GetLabels(), func(label *common.KeyValueItem) *kv.KV {
		return &kv.KV{Key: label.Key, Value: label.Value}
	})
	annotations := make(map[string]string, 2)
	annotations[cnst.AnnotationKeySummary] = request.GetAnnotations().GetSummary()
	annotations[cnst.AnnotationKeyDescription] = request.GetAnnotations().GetDescription()
	return &bo.SaveTeamMetricStrategyParams{
		StrategyID:  request.GetStrategyId(),
		Expr:        request.GetExpr(),
		Labels:      labels,
		Annotations: annotations,
		Datasource:  request.GetDatasource(),
	}
}

func ToSaveTeamMetricStrategyLevelParams(request *palace.SaveTeamMetricStrategyLevelRequest) *bo.SaveTeamMetricStrategyLevelParams {
	if validate.IsNil(request) {
		panic("SaveTeamMetricStrategyLevelRequest is nil")
	}
	return &bo.SaveTeamMetricStrategyLevelParams{
		SampleMode:            vobj.SampleMode(request.GetSampleMode()),
		Total:                 request.GetTotal(),
		Condition:             vobj.ConditionMetric(request.GetCondition()),
		Values:                request.GetValues(),
		LabelNotices:          slices.Map(request.GetLabelReceiverRoutes(), ToLabelNoticeParams),
		Duration:              request.GetDuration(),
		AlarmPages:            request.GetAlarmPages(),
		StrategyMetricID:      request.GetStrategyMetricId(),
		StrategyMetricLevelID: request.GetStrategyMetricLevelId(),
		LevelID:               request.GetLevelId(),
		ReceiverRoutesIds:     request.GetReceiverRoutes(),
	}
}

func ToListTeamMetricStrategyLevelsParams(request *palace.TeamMetricStrategyLevelListRequest) *bo.ListTeamMetricStrategyLevelsParams {
	if validate.IsNil(request) {
		panic("ListTeamMetricStrategyLevelsRequest is nil")
	}
	return &bo.ListTeamMetricStrategyLevelsParams{
		PaginationRequest: ToPaginationRequest(request.GetPagination()),
		StrategyMetricID:  request.GetStrategyMetricId(),
		LevelId:           request.GetLevelId(),
	}
}

func ToUpdateTeamMetricStrategyLevelStatusParams(request *palace.UpdateTeamMetricStrategyLevelStatusRequest) *bo.UpdateTeamMetricStrategyLevelStatusParams {
	if validate.IsNil(request) {
		panic("UpdateTeamMetricStrategyLevelStatusRequest is nil")
	}
	return &bo.UpdateTeamMetricStrategyLevelStatusParams{
		StrategyMetricLevelID: request.GetStrategyMetricLevelId(),
		Status:                vobj.GlobalStatus(request.GetStatus()),
	}
}

func ToTeamMetricStrategyRuleItems(rules []do.StrategyMetricRule) []*common.TeamStrategyMetricLevelItem {
	return slices.MapFilter(rules, func(rule do.StrategyMetricRule) (*common.TeamStrategyMetricLevelItem, bool) {
		if validate.IsNil(rule) {
			return nil, false
		}
		return ToTeamMetricStrategyRuleItem(rule), true
	})
}

func ToTeamMetricStrategyRuleItem(rule do.StrategyMetricRule) *common.TeamStrategyMetricLevelItem {
	if validate.IsNil(rule) {
		return nil
	}
	return &common.TeamStrategyMetricLevelItem{
		StrategyMetricLevelId: rule.GetID(),
		StrategyMetricId:      rule.GetStrategyMetricID(),
		Level:                 ToDictItem(rule.GetLevel()),
		SampleMode:            common.SampleMode(rule.GetSampleMode()),
		Condition:             common.ConditionMetric(rule.GetCondition()),
		Total:                 rule.GetTotal(),
		Values:                rule.GetValues(),
		Duration:              rule.GetDuration(),
		Status:                common.GlobalStatus(rule.GetStatus().GetValue()),
		AlarmPages:            ToDictItems(rule.GetAlarmPages()),
		ReceiverRoutes:        ToNoticeGroupItems(rule.GetNotices()),
		LabelReceiverRoutes:   ToLabelNoticeItems(rule.GetLabelNotices()),
	}
}

func ToLabelNoticeParams(request *common.LabelNotices) *bo.LabelNoticeParams {
	if validate.IsNil(request) {
		panic("LabelReceiverRoutes is nil")
	}
	return &bo.LabelNoticeParams{
		Key:               request.GetKey(),
		Value:             request.GetValue(),
		ReceiverRoutesIds: request.GetReceiverRoutes(),
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
		Base:                 ToTeamStrategyItem(strategy.GetStrategy()),
		StrategyMetricId:     strategy.GetID(),
		Expr:                 strategy.GetExpr(),
		Labels:               ToKeyValueItems(strategy.GetLabels()),
		Annotations:          ToAnnotationsItem(strategy.GetAnnotations()),
		StrategyMetricLevels: ToTeamMetricStrategyItemLevels(strategy.GetRules()),
		Datasource:           ToTeamMetricDatasourceItems(strategy.GetDatasourceList()),
		Creator:              ToUserBaseItem(strategy.GetCreator()),
	}
}

func ToKeyValueItems(labels []*kv.KV) []*common.KeyValueItem {
	return slices.Map(labels, ToKeyValueItem)
}

func ToKeyValueItem(label *kv.KV) *common.KeyValueItem {
	return &common.KeyValueItem{Key: label.Key, Value: label.Value}
}

func ToAnnotationsItem(annotations kv.StringMap) *common.AnnotationsItem {
	return &common.AnnotationsItem{
		Summary:     annotations[cnst.AnnotationKeySummary],
		Description: annotations[cnst.AnnotationKeyDescription],
	}
}

func ToTeamMetricStrategyItems(strategies []do.StrategyMetric) []*common.TeamStrategyMetricItem {
	return slices.Map(strategies, ToTeamMetricStrategyItem)
}

func ToTeamMetricStrategyItemLevel(rule do.StrategyMetricRule) *common.TeamStrategyMetricLevelItem {
	if validate.IsNil(rule) {
		return nil
	}
	return &common.TeamStrategyMetricLevelItem{
		StrategyMetricLevelId: rule.GetID(),
		StrategyMetricId:      rule.GetStrategyMetricID(),
		Level:                 ToDictItem(rule.GetLevel()),
		SampleMode:            common.SampleMode(rule.GetSampleMode()),
		Condition:             common.ConditionMetric(rule.GetCondition()),
		Total:                 rule.GetTotal(),
		Values:                rule.GetValues(),
		Duration:              rule.GetDuration(),
		Status:                common.GlobalStatus(rule.GetStatus().GetValue()),
		AlarmPages:            ToDictItems(rule.GetAlarmPages()),
		ReceiverRoutes:        ToNoticeGroupItems(rule.GetNotices()),
		LabelReceiverRoutes:   ToLabelNoticeItems(rule.GetLabelNotices()),
	}
}

func ToTeamMetricStrategyItemLevels(levels []do.StrategyMetricRule) []*common.TeamStrategyMetricLevelItem {
	return slices.MapFilter(levels, func(level do.StrategyMetricRule) (*common.TeamStrategyMetricLevelItem, bool) {
		if validate.IsNil(level) {
			return nil, false
		}
		return ToTeamMetricStrategyItemLevel(level), true
	})
}

func ToLabelNoticeItems(labelNotices []do.StrategyMetricRuleLabelNotice) []*common.StrategyMetricLevelLabelNotice {
	return slices.MapFilter(labelNotices, func(labelNotice do.StrategyMetricRuleLabelNotice) (*common.StrategyMetricLevelLabelNotice, bool) {
		if validate.IsNil(labelNotice) {
			return nil, false
		}
		return ToLabelNoticeItem(labelNotice), true
	})
}

func ToLabelNoticeItem(labelNotice do.StrategyMetricRuleLabelNotice) *common.StrategyMetricLevelLabelNotice {
	if validate.IsNil(labelNotice) {
		return nil
	}
	return &common.StrategyMetricLevelLabelNotice{
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
