package builder

import (
	"context"
	"time"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"google.golang.org/protobuf/types/known/durationpb"
)

var _ IStrategyModuleBuilder = (*strategyModuleBuilder)(nil)

type (
	strategyModuleBuilder struct {
		ctx context.Context
	}

	// IStrategyModuleBuilder 策略模块构造器
	IStrategyModuleBuilder interface {
		// WithCreateStrategyGroupRequest 设置创建策略组请求参数
		WithCreateStrategyGroupRequest(*strategyapi.CreateStrategyGroupRequest) ICreateStrategyGroupRequestBuilder
		// WithDeleteStrategyGroupRequest 设置删除策略组请求参数
		WithDeleteStrategyGroupRequest(*strategyapi.DeleteStrategyGroupRequest) IDeleteStrategyGroupRequestBuilder
		// WithListStrategyGroupRequest 设置获取策略组列表请求参数
		WithListStrategyGroupRequest(*strategyapi.ListStrategyGroupRequest) IListStrategyGroupRequestBuilder
		// WithUpdateStrategyGroupRequest 设置更新策略组请求参数
		WithUpdateStrategyGroupRequest(*strategyapi.UpdateStrategyGroupRequest) IUpdateStrategyGroupRequestBuilder
		// WithUpdateStrategyGroupStatusRequest 设置更新策略组状态请求参数
		WithUpdateStrategyGroupStatusRequest(*strategyapi.UpdateStrategyGroupStatusRequest) IUpdateStrategyGroupStatusRequestBuilder
		// DoStrategyGroupBuilder 策略组条目构造器
		DoStrategyGroupBuilder() IDoStrategyGroupBuilder
		// WithCreateStrategyRequest 设置创建策略请求参数
		WithCreateStrategyRequest(*strategyapi.CreateStrategyRequest) ICreateStrategyRequestBuilder
		// WithUpdateStrategyRequest 设置更新策略请求参数
		WithUpdateStrategyRequest(*strategyapi.UpdateStrategyRequest) IUpdateStrategyRequestBuilder
		// WithListStrategyRequest 设置获取策略列表请求参数
		WithListStrategyRequest(*strategyapi.ListStrategyRequest) IListStrategyRequestBuilder
		// WithUpdateStrategyStatusRequest 设置更新策略状态请求参数
		WithUpdateStrategyStatusRequest(*strategyapi.UpdateStrategyStatusRequest) IUpdateStrategyStatusRequestBuilder
		// DoStrategyBuilder 策略条目构造器
		DoStrategyBuilder() IDoStrategyBuilder
		// DoStrategyLevelBuilder 策略等级条目构造器
		DoStrategyLevelBuilder() IDoStrategyLevelBuilder
		// WithCreateTemplateStrategyRequest 设置创建模板策略请求参数
		WithCreateTemplateStrategyRequest(*strategyapi.CreateTemplateStrategyRequest) ICreateTemplateStrategyRequestBuilder
		// WithUpdateTemplateStrategyRequest 设置更新模板策略请求参数
		WithUpdateTemplateStrategyRequest(*strategyapi.UpdateTemplateStrategyRequest) IUpdateTemplateStrategyRequestBuilder
		// WithListTemplateStrategyRequest 设置获取模板策略列表请求参数
		WithListTemplateStrategyRequest(*strategyapi.ListTemplateStrategyRequest) IListTemplateStrategyRequestBuilder
		// WithUpdateTemplateStrategyStatusRequest 设置更新模板策略状态请求参数
		WithUpdateTemplateStrategyStatusRequest(*strategyapi.UpdateTemplateStrategyStatusRequest) IUpdateTemplateStrategyStatusRequestBuilder
		// DoTemplateStrategyBuilder 模板策略条目构造器
		DoTemplateStrategyBuilder() IDoTemplateStrategyBuilder
		// APIMutationStrategyLevelTemplateItems 转换为API对象
		APIMutationStrategyLevelTemplateItems() IMutationStrategyLevelTemplateBuilder
		// APIMutationStrategyLevelItems 转换为API对象
		APIMutationStrategyLevelItems() IMutationStrategyLevelBuilder
		// DoStrategyLevelTemplateBuilder 策略等级模板条目构造器
		DoStrategyLevelTemplateBuilder() IDoStrategyLevelTemplateBuilder
		// BoStrategyBuilder 策略业务对象构造器
		BoStrategyBuilder() IBoStrategyBuilder
		// BoStrategyDomainBuilder 策略域名业务对象构造器
		BoStrategyDomainBuilder() IBoStrategyDomainBuilder
		// BoStrategyEndpointBuilder 策略端点业务对象构造器
		BoStrategyEndpointBuilder() IBoStrategyEndpointBuilder
		// BoStrategyPingBuilder 策略Ping业务对象构造器
		BoStrategyPingBuilder() IBoStrategyPingBuilder
	}

	// IBoStrategyDomainBuilder 策略域名业务对象构造器
	IBoStrategyDomainBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bo.StrategyDomain) *api.DomainStrategyItem
	}

	boStrategyDomainBuilder struct {
		ctx context.Context
	}

	// IBoStrategyEndpointBuilder 策略端点业务对象构造器
	IBoStrategyEndpointBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bo.StrategyEndpoint) *api.HttpStrategyItem
	}

	boStrategyEndpointBuilder struct {
		ctx context.Context
	}

	// IBoStrategyPingBuilder 策略Ping业务对象构造器
	IBoStrategyPingBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bo.StrategyPing) *api.PingStrategyItem
	}

	boStrategyPingBuilder struct {
		ctx context.Context
	}

	// ICreateStrategyGroupRequestBuilder 创建策略组请求参数构造器
	ICreateStrategyGroupRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateStrategyGroupParams
	}

	createStrategyGroupRequestBuilder struct {
		ctx context.Context
		*strategyapi.CreateStrategyGroupRequest
	}

	// IDeleteStrategyGroupRequestBuilder 删除策略组请求参数构造器
	IDeleteStrategyGroupRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.DelStrategyGroupParams
	}

	deleteStrategyGroupRequestBuilder struct {
		ctx context.Context
		*strategyapi.DeleteStrategyGroupRequest
	}

	// IListStrategyGroupRequestBuilder 获取策略组列表请求参数构造器
	IListStrategyGroupRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryStrategyGroupListParams
	}

	listStrategyGroupRequestBuilder struct {
		ctx context.Context
		*strategyapi.ListStrategyGroupRequest
	}

	// IUpdateStrategyGroupRequestBuilder 更新策略组请求参数构造器
	IUpdateStrategyGroupRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateStrategyGroupParams
	}

	updateStrategyGroupRequestBuilder struct {
		ctx context.Context
		*strategyapi.UpdateStrategyGroupRequest
	}

	// IUpdateStrategyGroupStatusRequestBuilder 更新策略组状态请求参数构造器
	IUpdateStrategyGroupStatusRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateStrategyGroupStatusParams
	}

	updateStrategyGroupStatusRequestBuilder struct {
		ctx context.Context
		*strategyapi.UpdateStrategyGroupStatusRequest
	}

	// ICreateStrategyRequestBuilder 创建策略请求参数构造器
	ICreateStrategyRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateStrategyParams
	}

	createStrategyRequestBuilder struct {
		ctx context.Context
		*strategyapi.CreateStrategyRequest
	}

	// IUpdateStrategyRequestBuilder 更新策略请求参数构造器
	IUpdateStrategyRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateStrategyParams
	}

	updateStrategyRequestBuilder struct {
		ctx context.Context
		*strategyapi.UpdateStrategyRequest
	}

	// IListStrategyRequestBuilder 获取策略列表请求参数构造器
	IListStrategyRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryStrategyListParams
	}

	listStrategyRequestBuilder struct {
		ctx context.Context
		*strategyapi.ListStrategyRequest
	}

	// IUpdateStrategyStatusRequestBuilder 更新策略状态请求参数构造器
	IUpdateStrategyStatusRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateStrategyStatusParams
	}

	updateStrategyStatusRequestBuilder struct {
		ctx context.Context
		*strategyapi.UpdateStrategyStatusRequest
	}

	// ICreateTemplateStrategyRequestBuilder 创建模板策略请求参数构造器
	ICreateTemplateStrategyRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateTemplateStrategyParams
	}

	createTemplateStrategyRequestBuilder struct {
		ctx context.Context
		*strategyapi.CreateTemplateStrategyRequest
	}

	// IUpdateTemplateStrategyRequestBuilder 更新模板策略请求参数构造器
	IUpdateTemplateStrategyRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateTemplateStrategyParams
	}

	updateTemplateStrategyRequestBuilder struct {
		ctx context.Context
		*strategyapi.UpdateTemplateStrategyRequest
	}

	// IListTemplateStrategyRequestBuilder 获取模板策略列表请求参数构造器
	IListTemplateStrategyRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryTemplateStrategyListParams
	}

	listTemplateStrategyRequestBuilder struct {
		ctx context.Context
		*strategyapi.ListTemplateStrategyRequest
	}

	// IUpdateTemplateStrategyStatusRequestBuilder 更新模板策略状态请求参数构造器
	IUpdateTemplateStrategyStatusRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateTemplateStrategyStatusParams
	}

	updateTemplateStrategyStatusRequestBuilder struct {
		ctx context.Context
		*strategyapi.UpdateTemplateStrategyStatusRequest
	}

	// IDoStrategyGroupBuilder 策略组条目构造器
	IDoStrategyGroupBuilder interface {
		// WithStrategyCountMap 设置策略计数映射
		WithStrategyCountMap(*bo.StrategyCountMap) IDoStrategyGroupBuilder
		// ToAPI 转换为API对象
		ToAPI(*bizmodel.StrategyGroup, ...map[uint32]*adminapi.UserItem) *adminapi.StrategyGroupItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.StrategyGroup) []*adminapi.StrategyGroupItem
		// ToSelect 转换为选择对象
		ToSelect(*bizmodel.StrategyGroup) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
		ToSelects([]*bizmodel.StrategyGroup) []*adminapi.SelectItem
	}

	doStrategyGroupBuilder struct {
		ctx              context.Context
		strategyCountMap *bo.StrategyCountMap
	}

	// IDoStrategyBuilder 策略条目构造器
	IDoStrategyBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bizmodel.Strategy, ...map[uint32]*adminapi.UserItem) *adminapi.StrategyItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.Strategy) []*adminapi.StrategyItem
		// ToSelect 转换为选择对象
		ToSelect(*bizmodel.Strategy) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
		ToSelects([]*bizmodel.Strategy) []*adminapi.SelectItem
		// ToBos 转换为业务对象列表
		ToBos(*bizmodel.Strategy) []*bo.Strategy
		// ToBosV2 转换为业务对象列表
		ToBosV2(*bizmodel.Strategy) []*bo.Strategy
		// ToBoMetrics 转换为业务对象列表
		ToBoMetrics(*bizmodel.Strategy) []*bo.Strategy
		// ToMQs 转换为MQ对象列表
		ToMQs(*bizmodel.Strategy) []*bo.Strategy
	}

	doStrategyBuilder struct {
		ctx                 context.Context
		strategyLevelDetail *bo.StrategyLevelDetailModel
	}

	// IDoTemplateStrategyBuilder 模板策略条目构造器
	IDoTemplateStrategyBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*model.StrategyTemplate, ...map[uint32]*adminapi.UserItem) *adminapi.StrategyTemplateItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*model.StrategyTemplate) []*adminapi.StrategyTemplateItem
		// ToSelect 转换为选择对象
		ToSelect(*model.StrategyTemplate) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
		ToSelects([]*model.StrategyTemplate) []*adminapi.SelectItem
	}

	doTemplateStrategyBuilder struct {
		ctx context.Context
	}

	// IDoStrategyLevelTemplateBuilder 策略等级模板条目构造器
	IDoStrategyLevelTemplateBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*model.StrategyLevelTemplate, ...map[uint32]*adminapi.UserItem) *adminapi.StrategyLevelTemplateItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*model.StrategyLevelTemplate) []*adminapi.StrategyLevelTemplateItem
	}

	doStrategyLevelTemplateBuilder struct {
		ctx context.Context
	}

	// IMutationStrategyLevelTemplateBuilder 策略等级模板条目构造器
	IMutationStrategyLevelTemplateBuilder interface {
		// WithStrategyTemplateID 设置策略模板ID
		WithStrategyTemplateID(uint32) IMutationStrategyLevelTemplateBuilder
		// ToBo 转换为业务对象
		ToBo(*strategyapi.MutationStrategyLevelTemplateItem) *bo.CreateStrategyLevelTemplate
		// ToBos 转换为业务对象列表
		ToBos([]*strategyapi.MutationStrategyLevelTemplateItem) []*bo.CreateStrategyLevelTemplate
	}

	mutationStrategyLevelTemplateBuilder struct {
		ctx                context.Context
		StrategyTemplateID uint32
	}

	// IMutationStrategyLevelBuilder 策略等级条目构造器
	IMutationStrategyLevelBuilder interface {
		// WithStrategyID 设置策略ID
		WithStrategyID(uint32) IMutationStrategyLevelBuilder
		// ToMetricBo 转换为业务对象
		ToMetricBo(*strategyapi.CreateStrategyMetricLevelRequest) *bo.CreateStrategyMetricLevel
		// ToMetricBos 转换为业务对象列表
		ToMetricBos([]*strategyapi.CreateStrategyMetricLevelRequest) []*bo.CreateStrategyMetricLevel
		// ToMQBo 转换为业务对象
		ToMQBo(*strategyapi.CreateStrategyEventLevelRequest) *bo.CreateStrategyEventLevel
		// ToMQBos 转换为业务对象列表
		ToMQBos([]*strategyapi.CreateStrategyEventLevelRequest) []*bo.CreateStrategyEventLevel
		// ToDomainBo 转换为领域对象
		ToDomainBo(*strategyapi.CreateStrategyDomainLevelRequest) *bo.CreateStrategyDomainLevel
		// ToDomainBos 转换为领域对象列表
		ToDomainBos([]*strategyapi.CreateStrategyDomainLevelRequest) []*bo.CreateStrategyDomainLevel
		// ToPortBo 转换为业务对象
		ToPortBo(*strategyapi.CreateStrategyPortLevelRequest) *bo.CreateStrategyPortLevel
		// ToPortBos 转换为业务对象列表
		ToPortBos([]*strategyapi.CreateStrategyPortLevelRequest) []*bo.CreateStrategyPortLevel
		// ToHTTPBo 转换为业务对象
		ToHTTPBo(*strategyapi.CreateStrategyHTTPLevelRequest) *bo.CreateStrategyHTTPLevel
		// ToHTTPBos 转换为业务对象列表
		ToHTTPBos([]*strategyapi.CreateStrategyHTTPLevelRequest) []*bo.CreateStrategyHTTPLevel
	}

	mutationStrategyLevelBuilder struct {
		ctx        context.Context
		StrategyID uint32
	}

	// IDoStrategyLevelBuilder 策略等级条目构造器
	IDoStrategyLevelBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bizmodel.StrategyMetricsLevel) *adminapi.StrategyMetricLevelItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.StrategyMetricsLevel) []*adminapi.StrategyMetricLevelItem
		// ToMqAPI 转换为API对象
		ToMqAPI(*bizmodel.StrategyMQLevel) *adminapi.StrategyMQLevelItem
		// ToMqAPIs 转换为API对象列表
		ToMqAPIs([]*bizmodel.StrategyMQLevel) []*adminapi.StrategyMQLevelItem
		// ToDomainAPI 转换为API对象
		ToDomainAPI(*bizmodel.StrategyDomain) *adminapi.StrategyDomainLevelItem
		// ToDomainAPIs 转换为API对象
		ToDomainAPIs([]*bizmodel.StrategyDomain) []*adminapi.StrategyDomainLevelItem
		// ToPortAPI 转换为API对象
		ToPortAPI(*bizmodel.StrategyPort) *adminapi.StrategyPortLevelItem
		// ToPortAPIs 转换为API对象列表
		ToPortAPIs([]*bizmodel.StrategyPort) []*adminapi.StrategyPortLevelItem
		// ToHTTPAPI 转换为API对象
		ToHTTPAPI(*bizmodel.StrategyHTTP) *adminapi.StrategyHTTPLevelItem
		// ToHTTPAPIs 转换为API对象列表
		ToHTTPAPIs([]*bizmodel.StrategyHTTP) []*adminapi.StrategyHTTPLevelItem
	}

	doStrategyLevelBuilder struct {
		ctx context.Context
	}

	// IBoStrategyBuilder 策略业务对象构造器
	IBoStrategyBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bo.Strategy) *api.MetricStrategyItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bo.Strategy) []*api.MetricStrategyItem
		// ToMqAPI 转换为mq API对象
		ToMqAPI(*bo.Strategy) *api.MQStrategyItem
	}

	boStrategyBuilder struct {
		ctx context.Context
	}
)

func (d *doStrategyLevelBuilder) ToDomainAPI(domain *bizmodel.StrategyDomain) *adminapi.StrategyDomainLevelItem {
	if types.IsNil(domain) {
		return nil
	}
	return &adminapi.StrategyDomainLevelItem{
		LevelId: domain.LevelID,
		Level:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(domain.Level),
		Status:  1,
		AlarmPages: NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(domain.AlarmPage, func(item *bizmodel.SysDict) imodel.IDict {
			return imodel.IDict(item)
		})),
		AlarmGroups:  nil,
		LabelNotices: nil,
		Id:           domain.ID,
		Threshold:    domain.Threshold,
		Condition:    api.Condition(domain.Condition),
	}
}

func (d *doStrategyLevelBuilder) ToDomainAPIs(domains []*bizmodel.StrategyDomain) []*adminapi.StrategyDomainLevelItem {
	return types.SliceTo(domains, func(domain *bizmodel.StrategyDomain) *adminapi.StrategyDomainLevelItem {
		return d.ToDomainAPI(domain)
	})
}

func (d *doStrategyLevelBuilder) ToPortAPI(port *bizmodel.StrategyPort) *adminapi.StrategyPortLevelItem {
	if types.IsNil(port) {
		return nil
	}

	return &adminapi.StrategyPortLevelItem{
		LevelId: port.LevelID,
		Level:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(port.Level),
		Status:  1,
		AlarmPages: NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(port.AlarmPage, func(item *bizmodel.SysDict) imodel.IDict {
			return imodel.IDict(item)
		})),
		AlarmGroups:  nil,
		LabelNotices: nil,
		Id:           0,
		Threshold:    port.Threshold,
		Port:         port.Port,
	}
}

func (d *doStrategyLevelBuilder) ToPortAPIs(ports []*bizmodel.StrategyPort) []*adminapi.StrategyPortLevelItem {
	return types.SliceTo(ports, func(port *bizmodel.StrategyPort) *adminapi.StrategyPortLevelItem {
		return d.ToPortAPI(port)
	})
}

func (d *doStrategyLevelBuilder) ToHTTPAPI(http *bizmodel.StrategyHTTP) *adminapi.StrategyHTTPLevelItem {
	if types.IsNil(http) {
		return nil
	}
	header := make(map[string]string, len(http.Headers))
	for _, item := range http.Headers {
		header[item.Name] = item.Value
	}
	return &adminapi.StrategyHTTPLevelItem{
		LevelId: http.LevelID,
		Level:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(http.Level),
		Status:  1,
		AlarmPages: NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(http.AlarmPages, func(item *bizmodel.SysDict) imodel.IDict {
			return imodel.IDict(item)
		})),
		AlarmGroups:           nil,
		LabelNotices:          nil,
		Id:                    0,
		StatusCode:            http.StatusCode,
		ResponseTime:          http.ResponseTime,
		Headers:               header,
		Body:                  http.Body,
		QueryParams:           http.QueryParams,
		Method:                http.Method.String(),
		StatusCodeCondition:   api.Condition(http.StatusCodeCondition),
		ResponseTimeCondition: api.Condition(http.ResponseTimeCondition),
	}
}

func (d *doStrategyLevelBuilder) ToHTTPAPIs(https []*bizmodel.StrategyHTTP) []*adminapi.StrategyHTTPLevelItem {
	return types.SliceTo(https, func(http *bizmodel.StrategyHTTP) *adminapi.StrategyHTTPLevelItem {
		return d.ToHTTPAPI(http)
	})
}

func (m *mutationStrategyLevelBuilder) ToDomainBo(request *strategyapi.CreateStrategyDomainLevelRequest) *bo.CreateStrategyDomainLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return &bo.CreateStrategyDomainLevel{
		Status:        vobj.Status(request.GetStatus()),
		LabelNotices:  NewParamsBuild(m.ctx).AlarmNoticeGroupModuleBuilder().APICreateStrategyLabelNoticeRequest().ToBos(request.GetLabelNotices()),
		AlarmGroupIds: request.GetAlarmGroupIds(),
		Condition:     vobj.Condition(request.GetCondition()),
		Threshold:     request.GetThreshold(),
		LevelID:       request.GetLevelId(),
		AlarmPageIds:  request.GetAlarmPageIds(),
	}
}

func (m *mutationStrategyLevelBuilder) ToDomainBos(requests []*strategyapi.CreateStrategyDomainLevelRequest) []*bo.CreateStrategyDomainLevel {
	if types.IsNil(requests) || types.IsNil(m) {
		return nil
	}
	return types.SliceTo(requests, func(request *strategyapi.CreateStrategyDomainLevelRequest) *bo.CreateStrategyDomainLevel {
		return m.ToDomainBo(request)
	})
}

func (m *mutationStrategyLevelBuilder) ToPortBo(request *strategyapi.CreateStrategyPortLevelRequest) *bo.CreateStrategyPortLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return &bo.CreateStrategyPortLevel{
		Status:        vobj.Status(request.GetStatus()),
		LabelNotices:  NewParamsBuild(m.ctx).AlarmNoticeGroupModuleBuilder().APICreateStrategyLabelNoticeRequest().ToBos(request.GetLabelNotices()),
		AlarmGroupIds: request.GetAlarmGroupIds(),
		Threshold:     request.GetThreshold(),
		Port:          request.GetPort(),
		LevelID:       request.GetLevelId(),
		AlarmPageIds:  request.GetAlarmPageIds(),
	}
}

func (m *mutationStrategyLevelBuilder) ToPortBos(requests []*strategyapi.CreateStrategyPortLevelRequest) []*bo.CreateStrategyPortLevel {
	if types.IsNil(requests) || types.IsNil(m) {
		return nil
	}
	return types.SliceTo(requests, func(request *strategyapi.CreateStrategyPortLevelRequest) *bo.CreateStrategyPortLevel {
		return m.ToPortBo(request)
	})
}

func (m *mutationStrategyLevelBuilder) ToHTTPBo(request *strategyapi.CreateStrategyHTTPLevelRequest) *bo.CreateStrategyHTTPLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}

	return &bo.CreateStrategyHTTPLevel{
		Status:                vobj.Status(request.GetStatus()),
		LabelNotices:          NewParamsBuild(m.ctx).AlarmNoticeGroupModuleBuilder().APICreateStrategyLabelNoticeRequest().ToBos(request.GetLabelNotices()),
		AlarmGroupIds:         request.GetAlarmGroupIds(),
		AlarmPageIds:          request.GetAlarmPageIds(),
		ResponseTime:          request.GetResponseTime(),
		StatusCode:            request.GetStatusCode(),
		Body:                  request.GetBody(),
		QueryParams:           request.GetQueryParams(),
		Method:                request.GetMethod(),
		StatusCodeCondition:   vobj.Condition(request.GetStatusCodeCondition()),
		ResponseTimeCondition: vobj.Condition(request.GetResponseTimeCondition()),
		Headers: types.SliceTo(request.GetHeaders(), func(item *strategyapi.HeaderItem) *bo.HeaderItem {
			return &bo.HeaderItem{
				Key:   item.GetKey(),
				Value: item.GetValue(),
			}
		}),
		LevelID: request.GetLevelId(),
	}
}

func (m *mutationStrategyLevelBuilder) ToHTTPBos(requests []*strategyapi.CreateStrategyHTTPLevelRequest) []*bo.CreateStrategyHTTPLevel {
	if types.IsNil(requests) || types.IsNil(m) {
		return nil
	}
	return types.SliceTo(requests, func(request *strategyapi.CreateStrategyHTTPLevelRequest) *bo.CreateStrategyHTTPLevel {
		return m.ToHTTPBo(request)
	})
}

func (b *boStrategyBuilder) ToMqAPI(strategy *bo.Strategy) *api.MQStrategyItem {
	if types.IsNil(strategy) || types.IsNil(strategy.MQLevel) || types.IsNil(b) {
		return nil
	}
	mqLevel := strategy.MQLevel
	item := &api.MQStrategyItem{
		StrategyType:     api.StrategyType(strategy.StrategyType),
		StrategyID:       strategy.ID,
		TeamID:           strategy.TeamID,
		Status:           api.Status(strategy.Status),
		Alert:            strategy.Alert,
		LevelId:          mqLevel.LevelID,
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		ReceiverGroupIDs: strategy.ReceiverGroupIDs,
		Value:            mqLevel.Value,
		Condition:        api.MQCondition(mqLevel.Condition),
		DataType:         api.MQDataType(mqLevel.MQDataType),
		Topic:            strategy.Expr,
		Datasource:       NewParamsBuild(b.ctx).DatasourceModuleBuilder().BoDatasourceBuilder().ToMqAPIs(strategy.Datasource),
		DataKey:          mqLevel.PathKey,
	}

	return item
}

func (d *doStrategyBuilder) ToBoMetrics(strategy *bizmodel.Strategy) []*bo.Strategy {
	if types.IsNil(strategy) || types.IsNil(d) || types.IsNil(d.strategyLevelDetail) {
		return nil
	}
	levelDetail := d.strategyLevelDetail

	strategyLevels := levelDetail.LevelMap[strategy.ID]
	strategyMetricsLevels := make([]*bizmodel.StrategyMetricsLevel, 0)
	for _, item := range strategyLevels {
		strategyMetricsLevels = append(strategyMetricsLevels, item.StrategyMetricsLevels...)
	}
	receiverGroupIDs := types.SliceTo(strategy.AlarmNoticeGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID })
	return types.SliceTo(strategyMetricsLevels, func(level *bizmodel.StrategyMetricsLevel) *bo.Strategy {
		receiverGroupIDs = append(receiverGroupIDs, types.SliceTo(level.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID })...)
		labelNotices := types.SliceTo(level.LabelNotices, func(notice *bizmodel.StrategyMetricsLabelNotice) *bo.LabelNotices {
			return &bo.LabelNotices{
				Key:              notice.Name,
				Value:            notice.Value,
				ReceiverGroupIDs: types.SliceTo(notice.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID }),
			}
		})
		return &bo.Strategy{
			ReceiverGroupIDs:           types.MergeSliceWithUnique(receiverGroupIDs),
			LabelNotices:               labelNotices,
			ID:                         strategy.ID,
			Alert:                      strategy.Name,
			Expr:                       strategy.Expr,
			MultiDatasourceSustainType: 0, // TODO 多数据源控制
			Labels:                     strategy.Labels,
			Annotations:                strategy.Annotations,
			Datasource:                 NewParamsBuild(d.ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToBos(strategy.Datasource),
			Status: types.Ternary(!strategy.Status.IsEnable() || strategy.GetDeletedAt() > 0 ||
				!level.Status.IsEnable(), vobj.StatusDisable, vobj.StatusEnable),
			TeamID: middleware.GetTeamID(d.ctx),
			MetricLevel: &bo.CreateStrategyMetricLevel{
				StrategyTemplateID: strategy.TemplateID,
				Duration:           level.Duration,
				Count:              level.Count,
				SustainType:        level.SustainType,
				Interval:           level.Interval,
				Condition:          level.Condition,
				Threshold:          level.Threshold,
				Status:             level.Status,
				LevelID:            level.LevelID,
				AlarmPageIds:       level.AlarmPageIds,
				AlarmGroupIds:      level.AlarmGroupIds,
				StrategyID:         strategy.ID,
				LabelNotices: types.SliceTo(level.LabelNotices, func(notice *bizmodel.StrategyMetricsLabelNotice) *bo.StrategyLabelNotice {
					return &bo.StrategyLabelNotice{
						Name:          notice.Name,
						Value:         notice.Value,
						AlarmGroupIds: types.SliceTo(notice.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID }),
					}
				}),
			},
			StrategyType: strategy.StrategyType,
		}
	})
}

func (d *doStrategyBuilder) ToMQs(strategy *bizmodel.Strategy) []*bo.Strategy {
	if types.IsNil(strategy) || types.IsNil(d) || types.IsNil(d.strategyLevelDetail) {
		return nil
	}
	levelDetail := d.strategyLevelDetail
	strategyLevels := levelDetail.LevelMap[strategy.ID]
	mqLevels := make([]*bizmodel.StrategyMQLevel, 0)
	for _, item := range strategyLevels {
		mqLevels = append(mqLevels, item.StrategyMQLevels...)
	}

	receiverGroupIDs := types.SliceTo(strategy.AlarmNoticeGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID })
	return types.SliceTo(mqLevels, func(level *bizmodel.StrategyMQLevel) *bo.Strategy {
		receiverGroupIDs = append(receiverGroupIDs, types.SliceTo(level.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID })...)
		return &bo.Strategy{
			ReceiverGroupIDs:           types.MergeSliceWithUnique(receiverGroupIDs),
			ID:                         strategy.ID,
			Alert:                      strategy.Name,
			Expr:                       strategy.Expr,
			MultiDatasourceSustainType: 0, // TODO 多数据源控制
			Labels:                     strategy.Labels,
			Annotations:                strategy.Annotations,
			StrategyType:               strategy.StrategyType,
			Datasource:                 NewParamsBuild(d.ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToBos(strategy.Datasource),
			Status: types.Ternary(!strategy.Status.IsEnable() || strategy.GetDeletedAt() > 0 ||
				!level.Status.IsEnable(), vobj.StatusDisable, vobj.StatusEnable),
			TeamID: middleware.GetTeamID(d.ctx),
			MQLevel: &bo.CreateStrategyEventLevel{
				Value:         level.Value,
				Condition:     level.Condition,
				MQDataType:    level.DataType,
				LevelID:       level.AlarmLevelID,
				Status:        level.Status,
				AlarmPageIds:  level.AlarmPageIds,
				AlarmGroupIds: level.AlarmGroupIds,
				StrategyID:    strategy.ID,
				PathKey:       level.PathKey,
			},
		}
	})
}

func (d *doStrategyBuilder) ToBosV2(strategy *bizmodel.Strategy) []*bo.Strategy {
	if types.IsNil(strategy) || types.IsNil(d) || types.IsNil(d.strategyLevelDetail) {
		return nil
	}
	switch strategy.StrategyType {
	case vobj.StrategyTypeMetric:
		return d.ToBoMetrics(strategy)
	case vobj.StrategyTypeMQ:
		return d.ToMQs(strategy)
	default:
		return nil
	}
}

func (d *doStrategyLevelBuilder) ToMqAPI(level *bizmodel.StrategyMQLevel) *adminapi.StrategyMQLevelItem {
	if types.IsNil(level) || types.IsNil(d) {
		return nil
	}

	return &adminapi.StrategyMQLevelItem{
		Threshold:    level.Value,
		Condition:    api.MQCondition(level.Condition),
		DataType:     api.MQDataType(level.DataType),
		LevelId:      level.AlarmLevelID,
		Level:        NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(level.AlarmLevel),
		AlarmPages:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(level.AlarmPage, func(item *bizmodel.SysDict) imodel.IDict { return item })),
		StrategyId:   level.StrategyID,
		AlarmGroups:  NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(level.AlarmGroups),
		Status:       api.Status(level.Status),
		PathKey:      level.PathKey,
		LabelNotices: nil,
	}
}

func (d *doStrategyLevelBuilder) ToMqAPIs(levels []*bizmodel.StrategyMQLevel) []*adminapi.StrategyMQLevelItem {
	if types.IsNil(d) || types.IsNil(levels) {
		return nil
	}
	return types.SliceTo(levels, func(level *bizmodel.StrategyMQLevel) *adminapi.StrategyMQLevelItem {
		return d.ToMqAPI(level)
	})
}

func (m *mutationStrategyLevelBuilder) ToMQBo(request *strategyapi.CreateStrategyEventLevelRequest) *bo.CreateStrategyEventLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return &bo.CreateStrategyEventLevel{
		Value:         request.GetThreshold(),
		Condition:     vobj.MQCondition(request.GetCondition()),
		MQDataType:    vobj.MQDataType(request.DataType),
		LevelID:       request.GetLevelId(),
		Status:        vobj.Status(request.GetStatus()),
		AlarmPageIds:  request.GetAlarmPageIds(),
		AlarmGroupIds: request.GetAlarmGroupIds(),
		StrategyID:    m.StrategyID,
		PathKey:       request.GetPathKey(),
	}
}

func (m *mutationStrategyLevelBuilder) ToMQBos(request []*strategyapi.CreateStrategyEventLevelRequest) []*bo.CreateStrategyEventLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return types.SliceTo(request, func(item *strategyapi.CreateStrategyEventLevelRequest) *bo.CreateStrategyEventLevel {
		return m.ToMQBo(item)
	})
}

func (b *boStrategyDomainBuilder) ToAPI(domain *bo.StrategyDomain) *api.DomainStrategyItem {
	if types.IsNil(domain) || types.IsNil(b) {
		return nil
	}
	return &api.DomainStrategyItem{
		StrategyID:       domain.ID,
		LevelId:          domain.LevelID,
		TeamID:           domain.TeamID,
		ReceiverGroupIDs: domain.ReceiverGroupIDs,
		Status:           api.Status(domain.Status),
		Labels:           domain.Labels.Map(),
		Annotations:      domain.Annotations,
		Threshold:        domain.Threshold,
		Domain:           domain.Domain,
		Alert:            domain.Alert,
		Timeout:          domain.Timeout,
		Interval:         durationpb.New(time.Duration(domain.Interval) * time.Second),
		Port:             domain.Port,
		StrategyType:     api.StrategyType(domain.Type),
	}
}

func (b *boStrategyEndpointBuilder) ToAPI(endpoint *bo.StrategyEndpoint) *api.HttpStrategyItem {
	if types.IsNil(endpoint) || types.IsNil(b) {
		return nil
	}
	return &api.HttpStrategyItem{
		StrategyType:     api.StrategyType(endpoint.Type),
		Url:              endpoint.URL,
		StrategyID:       endpoint.ID,
		LevelId:          endpoint.LevelID,
		TeamID:           endpoint.TeamID,
		ReceiverGroupIDs: endpoint.ReceiverGroupIDs,
		Status:           api.Status(endpoint.Status),
		Labels:           endpoint.Labels.Map(),
		Annotations:      endpoint.Annotations,
		Threshold:        endpoint.Threshold,
		Timeout:          endpoint.Timeout,
		Interval:         durationpb.New(time.Duration(endpoint.Interval) * time.Second),
		StatusCodes:      endpoint.StatusCode,
		Alert:            endpoint.Alert,
		Headers:          endpoint.Headers,
		Body:             endpoint.Body,
		Method:           endpoint.Method.String(),
	}
}

func (b *boStrategyPingBuilder) ToAPI(ping *bo.StrategyPing) *api.PingStrategyItem {
	if types.IsNil(ping) || types.IsNil(b) {
		return nil
	}
	return &api.PingStrategyItem{
		StrategyType:     api.StrategyType(ping.Type),
		StrategyID:       ping.ID,
		TeamID:           ping.TeamID,
		Status:           api.Status(ping.Status),
		Alert:            ping.Alert,
		Interval:         durationpb.New(time.Duration(ping.Interval) * time.Second),
		LevelId:          ping.LevelID,
		Timeout:          ping.Timeout,
		Labels:           ping.Labels.Map(),
		Annotations:      ping.Annotations,
		ReceiverGroupIDs: ping.ReceiverGroupIDs,
		Address:          ping.Address,
		TotalCount:       ping.TotalPackets,
		SuccessCount:     ping.SuccessPackets,
		LossRate:         ping.LossRate,
		AvgDelay:         ping.AvgDelay,
		MaxDelay:         ping.MaxDelay,
		MinDelay:         ping.MinDelay,
		StdDev:           ping.StdDevDelay,
	}
}

func (s *strategyModuleBuilder) BoStrategyDomainBuilder() IBoStrategyDomainBuilder {
	return &boStrategyDomainBuilder{ctx: s.ctx}
}

func (s *strategyModuleBuilder) BoStrategyEndpointBuilder() IBoStrategyEndpointBuilder {
	return &boStrategyEndpointBuilder{ctx: s.ctx}
}

func (s *strategyModuleBuilder) BoStrategyPingBuilder() IBoStrategyPingBuilder {
	return &boStrategyPingBuilder{ctx: s.ctx}
}

func (d *doStrategyBuilder) ToBos(strategy *bizmodel.Strategy) []*bo.Strategy {
	if types.IsNil(strategy) || types.IsNil(d) || types.IsNil(d.strategyLevelDetail) {
		return nil
	}

	levelDetail := d.strategyLevelDetail
	levels := levelDetail.LevelMap[strategy.ID]
	metricsLevels := make([]*bizmodel.StrategyMetricsLevel, 0)
	for _, item := range levels {
		metricsLevels = append(metricsLevels, item.StrategyMetricsLevels...)
	}
	receiverGroupIDs := types.SliceTo(strategy.AlarmNoticeGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID })
	return types.SliceTo(metricsLevels, func(level *bizmodel.StrategyMetricsLevel) *bo.Strategy {
		receiverGroupIDs = append(receiverGroupIDs, types.SliceTo(level.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID })...)
		labelNotices := types.SliceTo(level.LabelNotices, func(notice *bizmodel.StrategyMetricsLabelNotice) *bo.LabelNotices {
			return &bo.LabelNotices{
				Key:              notice.Name,
				Value:            notice.Value,
				ReceiverGroupIDs: types.SliceTo(notice.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 { return group.ID }),
			}
		})
		return &bo.Strategy{
			ReceiverGroupIDs:           types.MergeSliceWithUnique(receiverGroupIDs),
			LabelNotices:               labelNotices,
			ID:                         strategy.ID,
			LevelID:                    level.LevelID,
			Alert:                      strategy.Name,
			Expr:                       strategy.Expr,
			For:                        level.Duration,
			Count:                      level.Count,
			SustainType:                level.SustainType,
			MultiDatasourceSustainType: 0, // TODO 多数据源控制
			Labels:                     strategy.Labels,
			Annotations:                strategy.Annotations,
			Interval:                   level.Interval,
			Datasource:                 NewParamsBuild(d.ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToBos(strategy.Datasource),
			Status: types.Ternary(!strategy.Status.IsEnable() || strategy.GetDeletedAt() > 0 ||
				!level.Status.IsEnable(), vobj.StatusDisable, vobj.StatusEnable),
			Condition:    level.Condition,
			Threshold:    level.Threshold,
			StrategyType: strategy.StrategyType,
			TeamID:       middleware.GetTeamID(d.ctx),
		}
	})
}

func (b *boStrategyBuilder) ToAPI(strategyItem *bo.Strategy) *api.MetricStrategyItem {
	if types.IsNil(strategyItem) || types.IsNil(strategyItem.MetricLevel) || types.IsNil(b) {
		return nil
	}

	metricLevel := strategyItem.MetricLevel

	return &api.MetricStrategyItem{
		Alert:                      strategyItem.Alert,
		Expr:                       strategyItem.Expr,
		For:                        durationpb.New(time.Duration(strategyItem.For) * time.Second),
		Count:                      strategyItem.Count,
		SustainType:                api.SustainType(metricLevel.SustainType),
		MultiDatasourceSustainType: api.MultiDatasourceSustainType(strategyItem.MultiDatasourceSustainType),
		Labels:                     strategyItem.Labels.Map(),
		Annotations:                strategyItem.Annotations.Map(),
		Interval:                   durationpb.New(time.Duration(metricLevel.Interval) * time.Second),
		Datasource:                 NewParamsBuild(b.ctx).DatasourceModuleBuilder().BoDatasourceBuilder().ToAPIs(strategyItem.Datasource),
		Id:                         strategyItem.ID,
		Status:                     api.Status(metricLevel.Status),
		Step:                       strategyItem.Step,
		Condition:                  api.Condition(strategyItem.Condition),
		Threshold:                  metricLevel.Threshold,
		LevelId:                    metricLevel.LevelID,
		TeamID:                     strategyItem.TeamID,
		ReceiverGroupIDs:           strategyItem.ReceiverGroupIDs,
		LabelNotices: types.SliceTo(metricLevel.LabelNotices, func(item *bo.StrategyLabelNotice) *api.LabelNotices {
			return &api.LabelNotices{
				Key:              item.Name,
				Value:            item.Value,
				ReceiverGroupIDs: item.AlarmGroupIds,
			}
		}),
		StrategyType: api.StrategyType(strategyItem.StrategyType),
	}
}
func (b *boStrategyBuilder) ToAPIs(strategies []*bo.Strategy) []*api.MetricStrategyItem {
	if types.IsNil(strategies) || types.IsNil(b) {
		return nil
	}

	return types.SliceTo(strategies, func(item *bo.Strategy) *api.MetricStrategyItem {
		return b.ToAPI(item)
	})
}

func (s *strategyModuleBuilder) BoStrategyBuilder() IBoStrategyBuilder {
	return &boStrategyBuilder{ctx: s.ctx}
}

func (m *mutationStrategyLevelBuilder) WithStrategyID(u uint32) IMutationStrategyLevelBuilder {
	if !types.IsNil(m) {
		m.StrategyID = u
	}

	return m
}

func (m *mutationStrategyLevelBuilder) ToMetricBo(request *strategyapi.CreateStrategyMetricLevelRequest) *bo.CreateStrategyMetricLevel {
	if types.IsNil(m) || types.IsNil(request) {
		return nil
	}

	return &bo.CreateStrategyMetricLevel{
		ID:                 request.GetId(),
		StrategyTemplateID: m.StrategyID,
		Duration:           request.GetDuration(),
		Count:              request.Count,
		SustainType:        vobj.Sustain(request.SustainType),
		Interval:           request.GetInterval(),
		Condition:          vobj.Condition(request.Condition),
		Threshold:          request.Threshold,
		LevelID:            request.LevelId,
		Status:             vobj.Status(request.Status),
		AlarmPageIds:       request.GetAlarmPageIds(),
		AlarmGroupIds:      request.GetAlarmGroupIds(),
		StrategyID:         m.StrategyID,
		LabelNotices:       NewParamsBuild(m.ctx).AlarmNoticeGroupModuleBuilder().APICreateStrategyLabelNoticeRequest().ToBos(request.GetLabelNotices()),
	}
}

func (m *mutationStrategyLevelBuilder) ToMetricBos(requests []*strategyapi.CreateStrategyMetricLevelRequest) []*bo.CreateStrategyMetricLevel {
	if types.IsNil(m) || types.IsNil(requests) {
		return nil
	}

	return types.SliceTo(requests, func(request *strategyapi.CreateStrategyMetricLevelRequest) *bo.CreateStrategyMetricLevel {
		return m.ToMetricBo(request)
	})
}

func (s *strategyModuleBuilder) APIMutationStrategyLevelItems() IMutationStrategyLevelBuilder {
	return &mutationStrategyLevelBuilder{ctx: s.ctx}
}

func (d *doStrategyLevelBuilder) ToAPI(level *bizmodel.StrategyMetricsLevel) *adminapi.StrategyMetricLevelItem {
	if types.IsNil(d) || types.IsNil(level) {
		return nil
	}

	return &adminapi.StrategyMetricLevelItem{
		Duration:     level.Duration,
		Count:        level.Count,
		SustainType:  api.SustainType(level.SustainType),
		Interval:     level.Interval,
		Status:       api.Status(level.Status),
		LevelId:      level.LevelID,
		Level:        NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(level.Level),
		AlarmPages:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(level.AlarmPage, func(item *bizmodel.SysDict) imodel.IDict { return item })),
		Threshold:    level.Threshold,
		Condition:    api.Condition(level.Condition),
		AlarmGroups:  NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(level.AlarmGroups),
		LabelNotices: NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoLabelNoticeBuilder().ToAPIs(level.LabelNotices),
	}
}

func (d *doStrategyLevelBuilder) ToAPIs(levels []*bizmodel.StrategyMetricsLevel) []*adminapi.StrategyMetricLevelItem {
	if types.IsNil(d) || types.IsNil(levels) {
		return nil
	}
	return types.SliceTo(levels, func(level *bizmodel.StrategyMetricsLevel) *adminapi.StrategyMetricLevelItem {
		return d.ToAPI(level)
	})
}

func (s *strategyModuleBuilder) DoStrategyLevelBuilder() IDoStrategyLevelBuilder {
	return &doStrategyLevelBuilder{ctx: s.ctx}
}

func (m *mutationStrategyLevelTemplateBuilder) ToBos(items []*strategyapi.MutationStrategyLevelTemplateItem) []*bo.CreateStrategyLevelTemplate {
	if types.IsNil(m) || types.IsNil(items) {
		return nil
	}

	return types.SliceTo(items, func(item *strategyapi.MutationStrategyLevelTemplateItem) *bo.CreateStrategyLevelTemplate {
		return m.ToBo(item)
	})
}

func (m *mutationStrategyLevelTemplateBuilder) WithStrategyTemplateID(u uint32) IMutationStrategyLevelTemplateBuilder {
	if !types.IsNil(m) {
		m.StrategyTemplateID = u
	}
	return m
}

func (m *mutationStrategyLevelTemplateBuilder) ToBo(item *strategyapi.MutationStrategyLevelTemplateItem) *bo.CreateStrategyLevelTemplate {
	if types.IsNil(m) || types.IsNil(item) {
		return nil
	}

	return &bo.CreateStrategyLevelTemplate{
		StrategyTemplateID: m.StrategyTemplateID,
		Duration:           types.NewDuration(item.GetDuration()),
		Count:              item.GetCount(),
		SustainType:        vobj.Sustain(item.GetSustainType()),
		Condition:          vobj.Condition(item.Condition),
		Threshold:          item.Threshold,
		LevelID:            item.GetLevelId(),
		Status:             vobj.StatusEnable,
	}
}

func (s *strategyModuleBuilder) APIMutationStrategyLevelTemplateItems() IMutationStrategyLevelTemplateBuilder {
	return &mutationStrategyLevelTemplateBuilder{ctx: s.ctx}
}

func (d *doStrategyLevelTemplateBuilder) ToAPI(template *model.StrategyLevelTemplate, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.StrategyLevelTemplateItem {
	if types.IsNil(d) || types.IsNil(template) {
		return nil
	}

	userMap := getUsers(d.ctx, userMaps, template.CreatorID)

	return &adminapi.StrategyLevelTemplateItem{
		Id:          template.ID,
		Duration:    template.Duration.GetDuration(),
		Count:       template.Count,
		SustainType: api.SustainType(template.SustainType),
		Status:      api.Status(template.Status),
		LevelId:     template.LevelID,
		Level:       NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(template.Level),
		Threshold:   template.Threshold,
		Condition:   api.Condition(template.Condition),
		StrategyId:  template.StrategyTemplateID,
		Creator:     userMap[template.CreatorID],
	}
}

func (d *doStrategyLevelTemplateBuilder) ToAPIs(templates []*model.StrategyLevelTemplate) []*adminapi.StrategyLevelTemplateItem {
	if types.IsNil(d) || types.IsNil(templates) {
		return nil
	}

	ids := types.SliceTo(templates, func(item *model.StrategyLevelTemplate) uint32 { return item.CreatorID })
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(templates, func(item *model.StrategyLevelTemplate) *adminapi.StrategyLevelTemplateItem {
		return d.ToAPI(item, userMap)
	})
}

func (s *strategyModuleBuilder) DoStrategyLevelTemplateBuilder() IDoStrategyLevelTemplateBuilder {
	return &doStrategyLevelTemplateBuilder{ctx: s.ctx}
}

func (d *doTemplateStrategyBuilder) ToAPI(template *model.StrategyTemplate, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.StrategyTemplateItem {
	if types.IsNil(d) || types.IsNil(template) {
		return nil
	}

	userMap := getUsers(d.ctx, userMaps, template.CreatorID)
	return &adminapi.StrategyTemplateItem{
		Id:          template.ID,
		Alert:       template.Alert,
		Expr:        template.Expr,
		Levels:      NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelTemplateBuilder().ToAPIs(template.StrategyLevelTemplates),
		Labels:      template.Labels.Map(),
		Annotations: template.Annotations.Map(),
		Status:      api.Status(template.Status),
		CreatedAt:   template.CreatedAt.String(),
		UpdatedAt:   template.UpdatedAt.String(),
		Remark:      template.Remark,
		Creator:     userMap[template.CreatorID],
		Categories:  NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(template.Categories, func(item *model.SysDict) imodel.IDict { return item })),
	}
}

func (d *doTemplateStrategyBuilder) ToAPIs(templates []*model.StrategyTemplate) []*adminapi.StrategyTemplateItem {
	if types.IsNil(d) || types.IsNil(templates) {
		return nil
	}

	ids := types.SliceTo(templates, func(item *model.StrategyTemplate) uint32 { return item.CreatorID })
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(templates, func(item *model.StrategyTemplate) *adminapi.StrategyTemplateItem {
		return d.ToAPI(item, userMap)
	})
}

func (d *doTemplateStrategyBuilder) ToSelect(template *model.StrategyTemplate) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(template) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    template.ID,
		Label:    template.Alert,
		Children: nil,
		Disabled: template.DeletedAt > 0 || !template.Status.IsEnable(),
		Extend:   &adminapi.SelectExtend{Remark: template.Remark},
	}
}

func (d *doTemplateStrategyBuilder) ToSelects(templates []*model.StrategyTemplate) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(templates) {
		return nil
	}

	return types.SliceTo(templates, func(item *model.StrategyTemplate) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (u *updateTemplateStrategyStatusRequestBuilder) ToBo() *bo.UpdateTemplateStrategyStatusParams {
	if types.IsNil(u) || types.IsNil(u.UpdateTemplateStrategyStatusRequest) {
		return nil
	}

	return &bo.UpdateTemplateStrategyStatusParams{
		IDs:    u.GetIds(),
		Status: vobj.Status(u.GetStatus()),
	}
}

func (l *listTemplateStrategyRequestBuilder) ToBo() *bo.QueryTemplateStrategyListParams {
	if types.IsNil(l) || types.IsNil(l.ListTemplateStrategyRequest) {
		return nil
	}

	return &bo.QueryTemplateStrategyListParams{
		Keyword: l.GetKeyword(),
		Page:    types.NewPagination(l.GetPagination()),
		Alert:   l.GetAlert(),
		Status:  vobj.Status(l.GetStatus()),
	}
}

func (u *updateTemplateStrategyRequestBuilder) ToBo() *bo.UpdateTemplateStrategyParams {
	if types.IsNil(u) || types.IsNil(u.UpdateTemplateStrategyRequest) {
		return nil
	}

	return &bo.UpdateTemplateStrategyParams{
		ID: u.GetId(),
		Data: &bo.CreateTemplateStrategyParams{
			Alert:                  u.GetAlert(),
			Expr:                   u.GetExpr(),
			Remark:                 u.GetRemark(),
			Labels:                 vobj.NewLabels(u.GetLabels()),
			Annotations:            vobj.NewAnnotations(u.GetAnnotations()),
			StrategyLevelTemplates: NewParamsBuild(u.ctx).StrategyModuleBuilder().APIMutationStrategyLevelTemplateItems().WithStrategyTemplateID(u.GetId()).ToBos(u.GetLevels()),
			CategoriesIDs:          u.GetCategoriesIds(),
		},
	}
}

func (c *createTemplateStrategyRequestBuilder) ToBo() *bo.CreateTemplateStrategyParams {
	if types.IsNil(c) || types.IsNil(c.CreateTemplateStrategyRequest) {
		return nil
	}

	return &bo.CreateTemplateStrategyParams{
		Alert:                  c.GetAlert(),
		Expr:                   c.GetExpr(),
		Remark:                 c.GetRemark(),
		Labels:                 vobj.NewLabels(c.GetLabels()),
		Annotations:            vobj.NewAnnotations(c.GetAnnotations()),
		StrategyLevelTemplates: NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelTemplateItems().ToBos(c.GetLevels()),
		CategoriesIDs:          c.GetCategoriesIds(),
	}
}

func (d *doStrategyBuilder) ToAPI(strategy *bizmodel.Strategy, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.StrategyItem {
	if types.IsNil(d) || types.IsNil(strategy) {
		return nil
	}

	userMap := getUsers(d.ctx, userMaps, strategy.CreatorID)
	strategyItem := &adminapi.StrategyItem{
		Name:              strategy.Name,
		Expr:              strategy.Expr,
		StrategyType:      api.StrategyType(strategy.StrategyType),
		Labels:            strategy.Labels.Map(),
		Annotations:       strategy.Annotations.Map(),
		Datasource:        NewParamsBuild(d.ctx).DatasourceModuleBuilder().DoDatasourceBuilder().ToAPIs(strategy.Datasource),
		Id:                strategy.ID,
		Status:            api.Status(strategy.Status),
		CreatedAt:         strategy.CreatedAt.String(),
		UpdatedAt:         strategy.UpdatedAt.String(),
		Remark:            strategy.Remark,
		GroupId:           strategy.GroupID,
		Group:             NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyGroupBuilder().ToAPI(strategy.Group),
		TemplateId:        strategy.TemplateID,
		TemplateSource:    api.TemplateSourceType(strategy.TemplateSource),
		Categories:        NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToAPIs(types.SliceTo(strategy.Categories, func(item *bizmodel.SysDict) imodel.IDict { return item })),
		AlarmNoticeGroups: NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(strategy.AlarmNoticeGroups),
		Creator:           userMap[strategy.CreatorID],
		MetricLevels:      NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToAPIs(strategy.Level.StrategyMetricsLevels),
		EventLevels:       NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToMqAPIs(strategy.Level.StrategyMQLevels),
		PortLevels:        NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToPortAPIs(strategy.Level.StrategyPorts),
		HttpLevels:        NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToHTTPAPIs(strategy.Level.StrategyHTTPs),
		DomainLevels:      NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToDomainAPIs(strategy.Level.StrategyDomains),
	}

	return strategyItem
}

func (d *doStrategyBuilder) ToAPIs(strategies []*bizmodel.Strategy) []*adminapi.StrategyItem {
	if types.IsNil(d) || types.IsNil(strategies) {
		return nil
	}

	ids := types.SliceTo(strategies, func(item *bizmodel.Strategy) uint32 { return item.CreatorID })
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(strategies, func(item *bizmodel.Strategy) *adminapi.StrategyItem {
		return d.ToAPI(item, userMap)
	})
}

func (d *doStrategyBuilder) ToSelect(strategy *bizmodel.Strategy) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(strategy) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    strategy.ID,
		Label:    strategy.Name,
		Children: nil,
		Disabled: strategy.DeletedAt > 0 || !strategy.Status.IsEnable(),
		Extend:   &adminapi.SelectExtend{Remark: strategy.Remark},
	}
}

func (d *doStrategyBuilder) ToSelects(strategies []*bizmodel.Strategy) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(strategies) {
		return nil
	}

	return types.SliceTo(strategies, func(item *bizmodel.Strategy) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (u *updateStrategyStatusRequestBuilder) ToBo() *bo.UpdateStrategyStatusParams {
	if types.IsNil(u) || types.IsNil(u.UpdateStrategyStatusRequest) {
		return nil
	}

	return &bo.UpdateStrategyStatusParams{
		Ids:    u.GetIds(),
		Status: vobj.Status(u.GetStatus()),
	}
}

func (l *listStrategyRequestBuilder) ToBo() *bo.QueryStrategyListParams {
	if types.IsNil(l) || types.IsNil(l.ListStrategyRequest) {
		return nil
	}

	return &bo.QueryStrategyListParams{
		Keyword:    l.GetKeyword(),
		Page:       types.NewPagination(l.GetPagination()),
		Alert:      "",
		Status:     vobj.Status(l.GetStatus()),
		SourceType: vobj.StrategyTemplateSource(l.GetDatasourceType()),
		StrategyTypes: types.SliceTo(l.GetStrategyTypes(), func(item api.StrategyType) vobj.StrategyType {
			return vobj.StrategyType(item)
		}),
	}
}

func (u *updateStrategyRequestBuilder) ToBo() *bo.UpdateStrategyParams {
	if types.IsNil(u) || types.IsNil(u.UpdateStrategyRequest) {
		return nil
	}

	return &bo.UpdateStrategyParams{
		ID:          u.GetId(),
		UpdateParam: NewParamsBuild(u.ctx).StrategyModuleBuilder().WithCreateStrategyRequest(u.GetData()).ToBo(),
	}
}

func (c *createStrategyRequestBuilder) ToBo() *bo.CreateStrategyParams {
	if types.IsNil(c) || types.IsNil(c.CreateStrategyRequest) {
		return nil
	}

	return &bo.CreateStrategyParams{
		GroupID:        c.GetGroupId(),
		TemplateID:     c.GetTemplateId(),
		Remark:         c.GetRemark(),
		Status:         vobj.StatusEnable,
		DatasourceIDs:  c.GetDatasourceIds(),
		TemplateSource: vobj.StrategyTemplateSource(c.GetSourceType()),
		Name:           c.GetName(),
		TeamID:         middleware.GetTeamID(c.ctx),
		Labels:         vobj.NewLabels(c.GetLabels()),
		Annotations:    vobj.NewAnnotations(c.GetAnnotations()),
		Expr:           c.GetExpr(),
		CategoriesIds:  c.GetCategoriesIds(),
		AlarmGroupIds:  c.GetAlarmGroupIds(),
		StrategyType:   vobj.StrategyType(c.GetStrategyType()),
		MetricLevels:   NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToMetricBos(c.GetStrategyMetricLevels()),
		EventLevels:    NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToMQBos(c.GetStrategyEventLevels()),
		DomainLevels:   NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToDomainBos(c.GetStrategyDomainLevels()),
		PortLevels:     NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToPortBos(c.GetStrategyPortLevels()),
		HTTPLevels:     NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToHTTPBos(c.GetStrategyHTTPLevels()),
	}
}

func (d *doStrategyGroupBuilder) WithStrategyCountMap(countMap *bo.StrategyCountMap) IDoStrategyGroupBuilder {
	if !types.IsNil(d) {
		d.strategyCountMap = countMap
	}
	return d
}

func (d *doStrategyGroupBuilder) ToAPI(group *bizmodel.StrategyGroup, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.StrategyGroupItem {
	if types.IsNil(d) || types.IsNil(group) {
		return nil
	}
	userMap := getUsers(d.ctx, userMaps, group.CreatorID)
	strategyCount := d.strategyCountMap
	return &adminapi.StrategyGroupItem{
		Id:                  group.ID,
		Name:                group.Name,
		Status:              api.Status(group.Status),
		CreatedAt:           group.CreatedAt.String(),
		UpdatedAt:           group.UpdatedAt.String(),
		Remark:              group.Remark,
		Creator:             userMap[group.CreatorID],
		CreatorId:           group.CreatorID,
		Strategies:          NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyBuilder().ToAPIs(group.Strategies),
		StrategyCount:       strategyCount.GetStrategyCountMap(group.ID),
		EnableStrategyCount: strategyCount.GetStrategyEnableMap(group.ID),
		Categories:          NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToAPIs(types.SliceTo(group.Categories, func(item *bizmodel.SysDict) imodel.IDict { return item })),
	}
}

func (d *doStrategyGroupBuilder) ToAPIs(groups []*bizmodel.StrategyGroup) []*adminapi.StrategyGroupItem {
	if types.IsNil(d) || types.IsNil(groups) {
		return nil
	}

	ids := types.SliceTo(groups, func(item *bizmodel.StrategyGroup) uint32 { return item.CreatorID })
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(groups, func(item *bizmodel.StrategyGroup) *adminapi.StrategyGroupItem {
		return d.ToAPI(item, userMap)
	})
}

func (d *doStrategyGroupBuilder) ToSelect(group *bizmodel.StrategyGroup) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(group) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    group.ID,
		Label:    group.Name,
		Children: nil,
		Disabled: group.DeletedAt > 0 || !group.Status.IsEnable(),
		Extend:   &adminapi.SelectExtend{Remark: group.Remark},
	}
}

func (d *doStrategyGroupBuilder) ToSelects(groups []*bizmodel.StrategyGroup) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(groups) {
		return nil
	}

	return types.SliceTo(groups, func(item *bizmodel.StrategyGroup) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (u *updateStrategyGroupStatusRequestBuilder) ToBo() *bo.UpdateStrategyGroupStatusParams {
	if types.IsNil(u) || types.IsNil(u.UpdateStrategyGroupStatusRequest) {
		return nil
	}

	return &bo.UpdateStrategyGroupStatusParams{
		IDs:    u.GetIds(),
		Status: vobj.Status(u.GetStatus()),
	}
}

func (u *updateStrategyGroupRequestBuilder) ToBo() *bo.UpdateStrategyGroupParams {
	if types.IsNil(u) || types.IsNil(u.UpdateStrategyGroupRequest) {
		return nil
	}

	return &bo.UpdateStrategyGroupParams{
		ID:          u.GetId(),
		UpdateParam: NewParamsBuild(u.ctx).StrategyModuleBuilder().WithCreateStrategyGroupRequest(u.GetUpdate()).ToBo(),
	}
}

func (l *listStrategyGroupRequestBuilder) ToBo() *bo.QueryStrategyGroupListParams {
	if types.IsNil(l) || types.IsNil(l.ListStrategyGroupRequest) {
		return nil
	}

	return &bo.QueryStrategyGroupListParams{
		Keyword:       l.GetKeyword(),
		Page:          types.NewPagination(l.GetPagination()),
		Status:        vobj.Status(l.GetStatus()),
		CategoriesIds: l.GetCategoriesIds(),
	}
}

func (d *deleteStrategyGroupRequestBuilder) ToBo() *bo.DelStrategyGroupParams {
	if types.IsNil(d) || types.IsNil(d.DeleteStrategyGroupRequest) {
		return nil
	}

	return &bo.DelStrategyGroupParams{
		ID: d.GetId(),
	}
}

func (c *createStrategyGroupRequestBuilder) ToBo() *bo.CreateStrategyGroupParams {
	if types.IsNil(c) || types.IsNil(c.CreateStrategyGroupRequest) {
		return nil
	}

	return &bo.CreateStrategyGroupParams{
		Name:          c.GetName(),
		Remark:        c.GetRemark(),
		Status:        vobj.Status(c.GetStatus()),
		CategoriesIds: c.GetCategoriesIds(),
	}
}

func (s *strategyModuleBuilder) WithCreateStrategyGroupRequest(request *strategyapi.CreateStrategyGroupRequest) ICreateStrategyGroupRequestBuilder {
	return &createStrategyGroupRequestBuilder{ctx: s.ctx, CreateStrategyGroupRequest: request}
}

func (s *strategyModuleBuilder) WithDeleteStrategyGroupRequest(request *strategyapi.DeleteStrategyGroupRequest) IDeleteStrategyGroupRequestBuilder {
	return &deleteStrategyGroupRequestBuilder{ctx: s.ctx, DeleteStrategyGroupRequest: request}
}

func (s *strategyModuleBuilder) WithListStrategyGroupRequest(request *strategyapi.ListStrategyGroupRequest) IListStrategyGroupRequestBuilder {
	return &listStrategyGroupRequestBuilder{ctx: s.ctx, ListStrategyGroupRequest: request}
}

func (s *strategyModuleBuilder) WithUpdateStrategyGroupRequest(request *strategyapi.UpdateStrategyGroupRequest) IUpdateStrategyGroupRequestBuilder {
	return &updateStrategyGroupRequestBuilder{ctx: s.ctx, UpdateStrategyGroupRequest: request}
}

func (s *strategyModuleBuilder) WithUpdateStrategyGroupStatusRequest(request *strategyapi.UpdateStrategyGroupStatusRequest) IUpdateStrategyGroupStatusRequestBuilder {
	return &updateStrategyGroupStatusRequestBuilder{ctx: s.ctx, UpdateStrategyGroupStatusRequest: request}
}

func (s *strategyModuleBuilder) DoStrategyGroupBuilder() IDoStrategyGroupBuilder {
	return &doStrategyGroupBuilder{ctx: s.ctx}
}

func (s *strategyModuleBuilder) WithCreateStrategyRequest(request *strategyapi.CreateStrategyRequest) ICreateStrategyRequestBuilder {
	return &createStrategyRequestBuilder{ctx: s.ctx, CreateStrategyRequest: request}
}

func (s *strategyModuleBuilder) WithUpdateStrategyRequest(request *strategyapi.UpdateStrategyRequest) IUpdateStrategyRequestBuilder {
	return &updateStrategyRequestBuilder{ctx: s.ctx, UpdateStrategyRequest: request}
}

func (s *strategyModuleBuilder) WithListStrategyRequest(request *strategyapi.ListStrategyRequest) IListStrategyRequestBuilder {
	return &listStrategyRequestBuilder{ctx: s.ctx, ListStrategyRequest: request}
}

func (s *strategyModuleBuilder) WithUpdateStrategyStatusRequest(request *strategyapi.UpdateStrategyStatusRequest) IUpdateStrategyStatusRequestBuilder {
	return &updateStrategyStatusRequestBuilder{ctx: s.ctx, UpdateStrategyStatusRequest: request}
}

func (s *strategyModuleBuilder) DoStrategyBuilder() IDoStrategyBuilder {
	return &doStrategyBuilder{ctx: s.ctx}
}

func (s *strategyModuleBuilder) WithCreateTemplateStrategyRequest(request *strategyapi.CreateTemplateStrategyRequest) ICreateTemplateStrategyRequestBuilder {
	return &createTemplateStrategyRequestBuilder{ctx: s.ctx, CreateTemplateStrategyRequest: request}
}

func (s *strategyModuleBuilder) WithUpdateTemplateStrategyRequest(request *strategyapi.UpdateTemplateStrategyRequest) IUpdateTemplateStrategyRequestBuilder {
	return &updateTemplateStrategyRequestBuilder{ctx: s.ctx, UpdateTemplateStrategyRequest: request}
}

func (s *strategyModuleBuilder) WithListTemplateStrategyRequest(request *strategyapi.ListTemplateStrategyRequest) IListTemplateStrategyRequestBuilder {
	return &listTemplateStrategyRequestBuilder{ctx: s.ctx, ListTemplateStrategyRequest: request}
}

func (s *strategyModuleBuilder) WithUpdateTemplateStrategyStatusRequest(request *strategyapi.UpdateTemplateStrategyStatusRequest) IUpdateTemplateStrategyStatusRequestBuilder {
	return &updateTemplateStrategyStatusRequestBuilder{ctx: s.ctx, UpdateTemplateStrategyStatusRequest: request}
}

func (s *strategyModuleBuilder) DoTemplateStrategyBuilder() IDoTemplateStrategyBuilder {
	return &doTemplateStrategyBuilder{ctx: s.ctx}
}
