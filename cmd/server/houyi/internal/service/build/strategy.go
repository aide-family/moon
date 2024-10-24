package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyBuilder 策略构建器
type StrategyBuilder struct {
	*api.MetricStrategyItem
}

// NewStrategyBuilder 创建策略构建器
func NewStrategyBuilder(strategyInfo *api.MetricStrategyItem) *StrategyBuilder {
	return &StrategyBuilder{
		MetricStrategyItem: strategyInfo,
	}
}

// ToBo 转换为业务对象
func (a *StrategyBuilder) ToBo() *bo.StrategyMetric {
	if types.IsNil(a) || types.IsNil(a.MetricStrategyItem) {
		return nil
	}
	strategyInfo := a.MetricStrategyItem
	return &bo.StrategyMetric{
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

func (a *DomainStrategyBuilder) ToBo() *bo.StrategyDomain {
	if types.IsNil(a) || types.IsNil(a.DomainStrategyItem) {
		return nil
	}
	return &bo.StrategyDomain{
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		ID:               a.GetStrategyID(),
		LevelID:          a.GetLevelID(),
		TeamID:           a.GetTeamID(),
		Status:           vobj.Status(a.GetStatus()),
		Alert:            a.GetAlert(),
		Threshold:        float64(a.GetThreshold()),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      a.GetAnnotations(),
		Domain:           a.GetDomain(),
		Timeout:          types.Ternary(a.GetTimeout() > 0, a.GetTimeout(), 5),
		Interval:         types.NewDuration(a.GetInterval()),
		Port:             a.GetPort(),
		Type:             vobj.StrategyType(a.GetStrategyType()),
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

func (a *HTTPStrategyBuilder) ToBo() *bo.StrategyEndpoint {
	if types.IsNil(a) || types.IsNil(a.HttpStrategyItem) {
		return nil
	}
	return &bo.StrategyEndpoint{
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

type PingStrategyBuilder struct {
	*api.PingStrategyItem
}

func NewPingStrategyBuilder(strategyInfo *api.PingStrategyItem) *PingStrategyBuilder {
	return &PingStrategyBuilder{
		PingStrategyItem: strategyInfo,
	}
}

func (a *PingStrategyBuilder) ToBo() *bo.StrategyPing {
	if types.IsNil(a) || types.IsNil(a.PingStrategyItem) {
		return nil
	}

	return &bo.StrategyPing{
		Type:             vobj.StrategyType(a.GetStrategyType()),
		StrategyID:       a.GetStrategyID(),
		TeamID:           a.GetTeamID(),
		Status:           vobj.Status(a.GetStatus()),
		Alert:            a.GetAlert(),
		Interval:         types.NewDuration(a.GetInterval()),
		LevelID:          a.GetLevelID(),
		Timeout:          a.GetTimeout(),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      a.GetAnnotations(),
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		Address:          a.GetAddress(),
		TotalPackets:     float64(a.GetTotalCount()),
		SuccessPackets:   float64(a.GetSuccessCount()),
		LossRate:         a.GetLossRate(),
		MinDelay:         float64(a.GetMinDelay()),
		MaxDelay:         float64(a.GetMaxDelay()),
		AvgDelay:         float64(a.GetAvgDelay()),
		StdDevDelay:      float64(a.GetStdDev()),
	}
}
