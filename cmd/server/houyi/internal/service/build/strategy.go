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
		ID:                         strategyInfo.GetStrategyID(),
		LevelID:                    strategyInfo.GetLevelId(),
		Alert:                      strategyInfo.GetAlert(),
		Expr:                       strategyInfo.GetExpr(),
		For:                        types.NewDuration(strategyInfo.GetFor()),
		Count:                      strategyInfo.GetCount(),
		SustainType:                vobj.Sustain(strategyInfo.GetSustainType()),
		MultiDatasourceSustainType: vobj.MultiDatasourceSustain(strategyInfo.GetMultiDatasourceSustainType()),
		Labels:                     vobj.NewLabels(strategyInfo.GetLabels()),
		Annotations:                vobj.NewAnnotations(strategyInfo.GetAnnotations()),
		Datasource: types.SliceTo(strategyInfo.GetDatasource(), func(ds *api.Datasource) *bo.Datasource {
			return NewDatasourceAPIBuilder(ds).ToBo()
		}),
		Status:    vobj.Status(strategyInfo.GetStatus()),
		Condition: vobj.Condition(strategyInfo.GetCondition()),
		Threshold: strategyInfo.GetThreshold(),
		TeamID:    strategyInfo.GetTeamID(),
	}
}

// DomainStrategyBuilder 域名策略构建器
type DomainStrategyBuilder struct {
	*api.DomainStrategyItem
}

// NewDomainStrategyBuilder 创建域名策略构建器
func NewDomainStrategyBuilder(strategyInfo *api.DomainStrategyItem) *DomainStrategyBuilder {
	return &DomainStrategyBuilder{
		DomainStrategyItem: strategyInfo,
	}
}

// ToBo 转换为业务对象
func (a *DomainStrategyBuilder) ToBo() *bo.StrategyDomain {
	if types.IsNil(a) || types.IsNil(a.DomainStrategyItem) {
		return nil
	}
	return &bo.StrategyDomain{
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		ID:               a.GetStrategyID(),
		LevelID:          a.GetLevelId(),
		TeamID:           a.GetTeamID(),
		Status:           vobj.Status(a.GetStatus()),
		Alert:            a.GetAlert(),
		Threshold:        float64(a.GetThreshold()),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      vobj.NewAnnotations(a.GetAnnotations()),
		Domain:           a.GetDomain(),
		Timeout:          types.Ternary(a.GetTimeout() > 0, a.GetTimeout(), 5),
		Interval:         types.NewDuration(a.GetInterval()),
		Port:             a.GetPort(),
		Type:             vobj.StrategyType(a.GetStrategyType()),
	}
}

// HTTPStrategyBuilder HTTP策略构建器
type HTTPStrategyBuilder struct {
	*api.HttpStrategyItem
}

// NewHTTPStrategyBuilder 创建HTTP策略构建器
func NewHTTPStrategyBuilder(strategyInfo *api.HttpStrategyItem) *HTTPStrategyBuilder {
	return &HTTPStrategyBuilder{
		HttpStrategyItem: strategyInfo,
	}
}

// ToBo 转换为业务对象
func (a *HTTPStrategyBuilder) ToBo() *bo.StrategyEndpoint {
	if types.IsNil(a) || types.IsNil(a.HttpStrategyItem) {
		return nil
	}
	return &bo.StrategyEndpoint{
		Type:             vobj.StrategyType(a.GetStrategyType()),
		URL:              a.GetUrl(),
		Timeout:          a.GetTimeout(),
		StatusCode:       a.GetStatusCodes(),
		Headers:          a.GetHeaders(),
		Body:             a.GetBody(),
		Method:           vobj.ToHTTPMethod(a.GetMethod()),
		Threshold:        float64(a.GetThreshold()),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      vobj.NewAnnotations(a.GetAnnotations()),
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		LabelNotices:     nil,
		TeamID:           a.GetTeamID(),
		Status:           vobj.Status(a.GetStatus()),
		Alert:            a.GetAlert(),
		Interval:         types.NewDuration(a.GetInterval()),
		LevelID:          a.GetLevelId(),
		ID:               a.GetStrategyID(),
	}
}

// PingStrategyBuilder Ping策略构建器
type PingStrategyBuilder struct {
	*api.PingStrategyItem
}

// NewPingStrategyBuilder 创建Ping策略构建器
func NewPingStrategyBuilder(strategyInfo *api.PingStrategyItem) *PingStrategyBuilder {
	return &PingStrategyBuilder{
		PingStrategyItem: strategyInfo,
	}
}

// ToBo 转换为业务对象
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
		LevelID:          a.GetLevelId(),
		Timeout:          a.GetTimeout(),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      vobj.NewAnnotations(a.GetAnnotations()),
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

// EventStrategyBuilder MQ策略构建器
type EventStrategyBuilder struct {
	*api.EventStrategyItem
}

// NewMQStrategyBuilder 创建MQ策略构建器
func NewMQStrategyBuilder(strategyInfo *api.EventStrategyItem) *EventStrategyBuilder {
	return &EventStrategyBuilder{
		EventStrategyItem: strategyInfo,
	}
}

// ToBo 转换为业务对象
func (a *EventStrategyBuilder) ToBo() *bo.StrategyEvent {
	if types.IsNil(a) || types.IsNil(a.EventStrategyItem) {
		return nil
	}
	return &bo.StrategyEvent{
		StrategyType:     vobj.StrategyType(a.GetStrategyType()),
		TeamID:           a.GetTeamID(),
		ReceiverGroupIDs: a.GetReceiverGroupIDs(),
		ID:               a.GetStrategyID(),
		LevelID:          a.GetLevelId(),
		Alert:            a.GetAlert(),
		Expr:             a.GetTopic(),
		Threshold:        a.GetValue(),
		Condition:        vobj.EventCondition(a.GetCondition()),
		DataType:         vobj.EventDataType(a.GetDataType()),
		DataKey:          a.GetDataKey(),
		Datasource:       NewMQDatasourceAPIBuilder(a.GetDatasource()...).ToBos(),
		Status:           vobj.Status(a.GetStatus()),
		Labels:           vobj.NewLabels(a.GetLabels()),
		Annotations:      vobj.NewAnnotations(a.GetAnnotations()),
	}
}
