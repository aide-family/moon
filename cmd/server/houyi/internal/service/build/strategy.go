package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyAPIBuilder 策略api构建器
type StrategyAPIBuilder struct {
	*api.Strategy
}

// NewStrategyAPIBuilder 创建策略api构建器
func NewStrategyAPIBuilder(strategy *api.Strategy) *StrategyAPIBuilder {
	return &StrategyAPIBuilder{
		Strategy: strategy,
	}
}

// ToBo 转换为业务对象
func (b *StrategyAPIBuilder) ToBo() *bo.Strategy {
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
		Labels:                     vobj.NewLabels(b.GetLabels()),
		Annotations:                b.GetAnnotations(),
		Interval:                   types.NewDuration(b.GetInterval()),
		Datasource: types.SliceTo(b.GetDatasource(), func(ds *api.Datasource) *bo.Datasource {
			return NewDatasourceAPIBuilder(ds).ToBo()
		}),
		Status: vobj.Status(b.GetStatus()),
	}
}

// StrategyBuilder 策略构建器
type StrategyBuilder struct {
	*api.Strategy
}

// NewStrategyBuilder 创建策略构建器
func NewStrategyBuilder(strategyInfo *api.Strategy) *StrategyBuilder {
	return &StrategyBuilder{
		Strategy: strategyInfo,
	}
}

// ToBo 转换为业务对象
func (a *StrategyBuilder) ToBo() *bo.Strategy {
	if types.IsNil(a) || types.IsNil(a.Strategy) {
		return nil
	}
	strategyInfo := a.Strategy
	return &bo.Strategy{
		ID:                         strategyInfo.GetId(),
		Alert:                      strategyInfo.GetAlert(),
		Expr:                       strategyInfo.GetExpr(),
		For:                        types.NewDuration(strategyInfo.GetFor()),
		Count:                      strategyInfo.GetCount(),
		SustainType:                vobj.Sustain(strategyInfo.GetSustainType()),
		MultiDatasourceSustainType: vobj.MultiDatasourceSustain(strategyInfo.GetMultiDatasourceSustainType()),
		Labels:                     vobj.NewLabels(strategyInfo.GetLabels()),
		Annotations:                strategyInfo.GetAnnotations(),
		Interval:                   types.NewDuration(strategyInfo.GetInterval()),
		Datasource: types.SliceTo(strategyInfo.GetDatasource(), func(ds *api.Datasource) *bo.Datasource {
			return NewDatasourceAPIBuilder(ds).ToBo()
		}),
		Status:    vobj.Status(strategyInfo.GetStatus()),
		Step:      strategyInfo.GetStep(),
		Condition: vobj.Condition(strategyInfo.GetCondition()),
		Threshold: strategyInfo.GetThreshold(),
	}
}
