package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToStrategyGroup(ctx context.Context, strategyGroup do.StrategyGroup) *team.StrategyGroup {
	if validate.IsNil(strategyGroup) {
		return nil
	}
	strategyGroupDo, ok := strategyGroup.(*team.StrategyGroup)
	if ok {
		strategyGroupDo.WithContext(ctx)
		strategyGroupDo.Strategies = ToStrategies(ctx, strategyGroup.GetStrategies())
		return strategyGroupDo
	}
	strategyGroupDo = &team.StrategyGroup{
		Name:       strategyGroup.GetName(),
		TeamModel:  ToTeamModel(ctx, strategyGroup),
		Remark:     strategyGroup.GetRemark(),
		Status:     vobj.GlobalStatusEnable,
		Strategies: ToStrategies(ctx, strategyGroup.GetStrategies()),
	}
	strategyGroupDo.WithContext(ctx)
	return strategyGroupDo
}

func ToStrategy(ctx context.Context, params do.Strategy) *team.Strategy {
	if validate.IsNil(params) {
		return nil
	}
	strategy, ok := params.(*team.Strategy)
	if ok {
		strategy.StrategyGroup = ToStrategyGroup(ctx, params.GetStrategyGroup())
		strategy.Notices = ToStrategyNotices(ctx, params.GetNotices())
		strategy.WithContext(ctx)
		return strategy
	}
	strategy = &team.Strategy{
		StrategyGroupID: params.GetStrategyGroupID(),
		Name:            params.GetName(),
		Remark:          params.GetRemark(),
		Status:          vobj.GlobalStatusEnable,
		StrategyType:    params.GetStrategyType(),
		Notices:         ToStrategyNotices(ctx, params.GetNotices()),
		TeamModel:       ToTeamModel(ctx, params),
		StrategyGroup:   ToStrategyGroup(ctx, params.GetStrategyGroup()),
	}
	strategy.WithContext(ctx)
	return strategy
}

func ToStrategies(ctx context.Context, params []do.Strategy) []*team.Strategy {
	if len(params) == 0 {
		return nil
	}
	return slices.MapFilter(params, func(v do.Strategy) (*team.Strategy, bool) {
		item := ToStrategy(ctx, v)
		return item, validate.IsNotNil(item)
	})
}

func ToStrategyMetric(ctx context.Context, params do.StrategyMetric) *team.StrategyMetric {
	if validate.IsNil(params) {
		return nil
	}
	if strategyMetric, ok := params.(*team.StrategyMetric); ok {
		strategyMetric.WithContext(ctx)
		strategyMetric.Strategy = ToStrategy(ctx, params.GetStrategy())
		for i := range strategyMetric.Datasource {
			strategyMetric.Datasource[i] = ToDatasourceMetric(ctx, strategyMetric.Datasource[i])
		}
		return strategyMetric
	}
	strategyMetricDo := &team.StrategyMetric{
		StrategyID:          params.GetID(),
		Expr:                params.GetExpr(),
		Labels:              params.GetLabels(),
		Annotations:         params.GetAnnotations(),
		TeamModel:           ToTeamModel(ctx, params),
		Strategy:            ToStrategy(ctx, params.GetStrategy()),
		StrategyMetricRules: ToStrategyMetricRules(ctx, params.GetRules()),
		Datasource:          ToDatasourceMetrics(ctx, params.GetDatasourceList()),
	}
	strategyMetricDo.WithContext(ctx)
	return strategyMetricDo
}

func ToStrategyMetricRules(ctx context.Context, params []do.StrategyMetricRule) []*team.StrategyMetricRule {
	return slices.MapFilter(params, func(v do.StrategyMetricRule) (*team.StrategyMetricRule, bool) {
		if validate.IsNil(v) {
			return nil, false
		}
		return ToStrategyMetricRule(ctx, v), true
	})
}

func ToStrategyMetricRule(ctx context.Context, params do.StrategyMetricRule) *team.StrategyMetricRule {
	if validate.IsNil(params) {
		return nil
	}
	if strategyMetricRule, ok := params.(*team.StrategyMetricRule); ok {
		strategyMetricRule.WithContext(ctx)
		return strategyMetricRule
	}
	strategyMetricRuleDo := &team.StrategyMetricRule{
		TeamModel:        ToTeamModel(ctx, params),
		StrategyMetricID: params.GetID(),
		StrategyMetric:   ToStrategyMetric(ctx, params.GetStrategyMetric()),
		LevelID:          params.GetLevelID(),
		Level:            ToDict(ctx, params.GetLevel()),
		SampleMode:       params.GetSampleMode(),
		Condition:        params.GetCondition(),
		Total:            params.GetTotal(),
		Values:           params.GetValues(),
		Duration:         params.GetDuration(),
		Status:           params.GetStatus(),
		Notices:          ToStrategyNotices(ctx, params.GetNotices()),
		LabelNotices:     ToStrategyMetricRuleLabelNotices(ctx, params.GetLabelNotices()),
		AlarmPages:       ToDicts(ctx, params.GetAlarmPages()),
	}
	strategyMetricRuleDo.WithContext(ctx)
	return strategyMetricRuleDo
}
