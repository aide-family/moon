package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

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
		ReceiverGroupIDs: strategyInfo.GetReceiverGroupIDs(),
		LabelNotices: types.SliceTo(strategyInfo.LabelNotices, func(item *api.LabelNotices) *bo.LabelNotices {
			return &bo.LabelNotices{
				Key:              item.GetKey(),
				Value:            item.GetValue(),
				ReceiverGroupIDs: item.GetReceiverGroupIDs(),
			}
		}),
		ID:                         strategyInfo.GetId(),
		LevelID:                    strategyInfo.GetLevelID(),
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
		TeamID:    strategyInfo.GetTeamID(),
	}
}

type DomainStrategyBuilder struct {
	*api.DomainStrategyItem
}

func NewDomainStrategyBuilder(strategyInfo *api.DomainStrategyItem) *DomainStrategyBuilder {
	return &DomainStrategyBuilder{
		DomainStrategyItem: strategyInfo,
	}
}

func (a *DomainStrategyBuilder) ToBo() *bo.DomainStrategy {
	if types.IsNil(a) || types.IsNil(a.DomainStrategyItem) {
		return nil
	}
	return &bo.DomainStrategy{
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		LabelNotices: types.SliceTo(a.GetLabelNotices(), func(item *api.LabelNotices) *bo.LabelNotices {
			return &bo.LabelNotices{
				Key:              item.GetKey(),
				Value:            item.GetValue(),
				ReceiverGroupIDs: item.GetReceiverGroupIDs(),
			}
		}),
		ID:          a.GetStrategyID(),
		LevelID:     a.GetLevelID(),
		TeamID:      a.GetTeamID(),
		Status:      vobj.Status(a.GetStatus()),
		Alert:       a.GetAlert(),
		Threshold:   float64(a.GetThreshold()),
		Labels:      vobj.NewLabels(a.GetLabels()),
		Annotations: a.GetAnnotations(),
		Domain:      a.GetDomain(),
		Timeout:     types.Ternary(a.GetTimeout() > 0, a.GetTimeout(), 5),
		Interval:    types.NewDuration(a.GetInterval()),
		Port:        a.GetPort(),
		Type:        vobj.StrategyType(a.GetStrategyType()),
	}
}

type HTTPStrategyBuilder struct {
	*api.HttpStrategyItem
}

func NewHTTPStrategyBuilder(strategyInfo *api.HttpStrategyItem) *HTTPStrategyBuilder {
	return &HTTPStrategyBuilder{
		HttpStrategyItem: strategyInfo,
	}
}

func (a *HTTPStrategyBuilder) ToBo() *bo.EndpointDurationStrategy {
	if types.IsNil(a) || types.IsNil(a.HttpStrategyItem) {
		return nil
	}
	return &bo.EndpointDurationStrategy{
		Type:             vobj.StrategyType(a.GetStrategyType()),
		Url:              a.GetUrl(),
		Timeout:          a.GetTimeout(),
		StatusCode:       a.GetStatusCodes(),
		Headers:          a.GetHeaders(),
		Body:             a.GetBody(),
		Method:           vobj.ToHTTPMethod(a.GetMethod()),
		Threshold:        float64(a.GetThreshold()),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      a.GetAnnotations(),
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		LabelNotices:     nil,
		TeamID:           a.GetTeamID(),
		Status:           vobj.Status(a.GetStatus()),
		Alert:            a.GetAlert(),
		Interval:         types.NewDuration(a.GetInterval()),
		LevelID:          a.GetLevelID(),
		ID:               a.GetStrategyID(),
	}
}
