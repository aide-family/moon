package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type StrategyBuilder struct {
	*bizmodel.Strategy
}

func NewStrategyBuilder(strategy *bizmodel.Strategy) *StrategyBuilder {
	return &StrategyBuilder{
		Strategy: strategy,
	}
}

// ToApi 转换为API层数据
func (b *StrategyBuilder) ToApi(ctx context.Context) *admin.Strategy {
	if types.IsNil(b) || types.IsNil(b.Strategy) {
		return nil
	}
	strategyLevels := types.SliceToWithFilter(b.StrategyLevel, func(level *bizmodel.StrategyLevel) (*admin.StrategyLevel, bool) {
		return NewStrategyLevelBuilder(level).ToApi(ctx), true
	})

	return &admin.Strategy{
		Name:        b.Name,
		Id:          b.ID,
		Expr:        b.Expr,
		Labels:      b.Labels.Map(),
		Annotations: b.Annotations,
		Datasource: types.SliceTo(b.Datasource, func(datasource *bizmodel.Datasource) *admin.Datasource {
			return NewDatasourceBuilder(datasource).ToApi(ctx)
		}),
		StrategyTemplateId: b.StrategyTemplateID,
		Levels:             strategyLevels,
		Status:             api.Status(b.Status),
		Step:               b.Step,
		SourceType:         api.TemplateSourceType(b.StrategyTemplateSource),
	}
}
