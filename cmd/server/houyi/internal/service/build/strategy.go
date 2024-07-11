package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type StrategyApiBuilder struct {
	*api.Strategy
}

func NewStrategyApiBuilder(strategy *api.Strategy) *StrategyApiBuilder {
	return &StrategyApiBuilder{
		Strategy: strategy,
	}
}

func (b *StrategyApiBuilder) ToBo() *bo.Strategy {
	if types.IsNil(b) || types.IsNil(b.Strategy) {
		return nil
	}
	return &bo.Strategy{
		ID:                         b.GetId(),
		Alert:                      b.GetAlert(),
		Expr:                       b.GetExpr(),
		For:                        types.NewDuration(b.GetFor()),
		Count:                      b.GetCount(),
		SustainType:                vobj.Sustain(b.GetSustainType()),
		MultiDatasourceSustainType: vobj.MultiDatasourceSustain(b.GetMultiDatasourceSustainType()),
		Labels:                     b.GetLabels(),
		Annotations:                b.GetAnnotations(),
		Interval:                   types.NewDuration(b.GetInterval()),
		Datasource: types.SliceTo(b.GetDatasource(), func(ds *api.Datasource) *bo.Datasource {
			return NewDatasourceApiBuilder(ds).ToBo()
		}),
		Status: vobj.Status(b.GetStatus()),
	}
}
