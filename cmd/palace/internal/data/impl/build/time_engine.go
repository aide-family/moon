package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/pkg/util/slices"
)

func ToTimeEngineRules(ctx context.Context, rules []do.TimeEngineRule) []*team.TimeEngineRule {
	return slices.Map(rules, func(v do.TimeEngineRule) *team.TimeEngineRule {
		return ToTimeEngineRule(ctx, v)
	})
}

func ToTimeEngineRule(ctx context.Context, rule do.TimeEngineRule) *team.TimeEngineRule {
	item := &team.TimeEngineRule{
		TeamModel: ToTeamModel(ctx, rule),
		Name:      rule.GetName(),
		Remark:    rule.GetRemark(),
		Status:    rule.GetStatus(),
		Rule:      rule.GetRules(),
		Type:      rule.GetType(),
		Engines:   ToTimeEngines(ctx, rule.GetTimeEngines()),
	}
	item.WithContext(ctx)
	return item
}

func ToTimeEngines(ctx context.Context, engines []do.TimeEngine) []*team.TimeEngine {
	return slices.Map(engines, func(v do.TimeEngine) *team.TimeEngine {
		return ToTimeEngine(ctx, v)
	})
}

func ToTimeEngine(ctx context.Context, engine do.TimeEngine) *team.TimeEngine {
	item := &team.TimeEngine{
		TeamModel: ToTeamModel(ctx, engine),
		Name:      engine.GetName(),
		Remark:    engine.GetRemark(),
		Status:    engine.GetStatus(),
		Rules:     ToTimeEngineRules(ctx, engine.GetRules()),
	}
	item.WithContext(ctx)
	return item
}
