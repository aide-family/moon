package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type StrategyLevelBuilder struct {
	*bizmodel.StrategyLevel
}

func NewStrategyLevelBuilder(strategyLevel *bizmodel.StrategyLevel) *StrategyLevelBuilder {
	return &StrategyLevelBuilder{
		StrategyLevel: strategyLevel,
	}
}

func (b *StrategyLevelBuilder) ToApi(ctx context.Context) *admin.StrategyLevel {
	if types.IsNil(b) || types.IsNil(b.StrategyLevel) {
		return nil
	}

	strategyLevel := &admin.StrategyLevel{
		Duration:    b.Duration.GetDuration(),
		Count:       b.Count,
		SustainType: api.SustainType(b.SustainType),
		Interval:    b.Interval.GetDuration(),
		Status:      api.Status(b.Status),
		Id:          b.ID,
		LevelId:     b.LevelID,
		Threshold:   b.Threshold,
		StrategyId:  b.StrategyID,
		Condition:   api.Condition(b.Condition),
	}
	return strategyLevel
}
