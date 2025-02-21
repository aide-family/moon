package builder

import (
	"context"
	"time"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	houyistrategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/label"
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
		// DoStrategyLevelsBuilder 策略等级条目构造器
		DoStrategyLevelsBuilder() IDoStrategyLevelsBuilder
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
		ToAPI(*bizmodel.StrategyGroup) *adminapi.StrategyGroupItem
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
		ToAPI(*bizmodel.Strategy) *adminapi.StrategyItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.Strategy) []*adminapi.StrategyItem
		// ToSelect 转换为选择对象
		ToSelect(*bizmodel.Strategy) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
		ToSelects([]*bizmodel.Strategy) []*adminapi.SelectItem
		// ToBos 转换为业务对象列表
		ToBos(*bizmodel.Strategy) []*bo.Strategy
	}

	doStrategyBuilder struct {
		ctx context.Context
	}

	// IDoTemplateStrategyBuilder 模板策略条目构造器
	IDoTemplateStrategyBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*model.StrategyTemplate) *adminapi.StrategyTemplateItem
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
		ToAPI(*model.StrategyLevelTemplate) *adminapi.StrategyLevelTemplateItem
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
		// ToEventBo 转换为业务对象
		ToEventBo(*strategyapi.CreateStrategyEventLevelRequest) *bo.CreateStrategyEventLevel
		// ToEventBos 转换为业务对象列表
		ToEventBos([]*strategyapi.CreateStrategyEventLevelRequest) []*bo.CreateStrategyEventLevel
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
		// ToMetricAPI 转换为API对象
		ToMetricAPI(*bizmodel.StrategyMetricLevel) *adminapi.StrategyMetricLevelItem
		// ToMetricAPIs 转换为API对象列表
		ToMetricAPIs([]*bizmodel.StrategyMetricLevel) []*adminapi.StrategyMetricLevelItem
		// ToEventAPI 转换为API对象
		ToEventAPI(*bizmodel.StrategyEventLevel) *adminapi.StrategyEventLevelItem
		// ToEventAPIs 转换为API对象列表
		ToEventAPIs([]*bizmodel.StrategyEventLevel) []*adminapi.StrategyEventLevelItem
		// ToDomainAPI 转换为API对象
		ToDomainAPI(*bizmodel.StrategyDomainLevel) *adminapi.StrategyDomainLevelItem
		// ToDomainAPIs 转换为API对象
		ToDomainAPIs([]*bizmodel.StrategyDomainLevel) []*adminapi.StrategyDomainLevelItem
		// ToPortAPI 转换为API对象
		ToPortAPI(*bizmodel.StrategyPortLevel) *adminapi.StrategyPortLevelItem
		// ToPortAPIs 转换为API对象列表
		ToPortAPIs([]*bizmodel.StrategyPortLevel) []*adminapi.StrategyPortLevelItem
		// ToHTTPLevelAPI 转换为API对象
		ToHTTPLevelAPI(*bizmodel.StrategyHTTPLevel) *adminapi.StrategyHTTPLevelItem
		// ToHTTPLevelAPIs 转换为API对象列表
		ToHTTPLevelAPIs([]*bizmodel.StrategyHTTPLevel) []*adminapi.StrategyHTTPLevelItem
	}

	doStrategyLevelBuilder struct {
		ctx context.Context
	}

	// IDoStrategyLevelsBuilder 策略等级条目构造器
	IDoStrategyLevelsBuilder interface {
		// ToMetricAPI 转换为API对象
		ToMetricAPI(*bizmodel.Strategy, *bizmodel.StrategyMetricLevel) *api.MetricStrategyItem
		// ToMetricAPIs 转换为API对象列表
		ToMetricAPIs(*bizmodel.Strategy, []*bizmodel.StrategyMetricLevel) []*api.MetricStrategyItem
		// ToEventAPI 转换为API对象
		ToEventAPI(*bizmodel.Strategy, *bizmodel.StrategyEventLevel) *api.EventStrategyItem
		// ToEventAPIs 转换为API对象列表
		ToEventAPIs(*bizmodel.Strategy, []*bizmodel.StrategyEventLevel) []*api.EventStrategyItem
		// ToDomainAPI 转换为API对象
		ToDomainAPI(*bizmodel.Strategy, *bizmodel.StrategyDomainLevel) *api.DomainStrategyItem
		// ToDomainAPIs 转换为API对象
		ToDomainAPIs(*bizmodel.Strategy, []*bizmodel.StrategyDomainLevel) []*api.DomainStrategyItem
		// ToPortAPI 转换为API对象
		ToPortAPI(*bizmodel.Strategy, *bizmodel.StrategyPortLevel) *api.DomainStrategyItem
		// ToPortAPIs 转换为API对象列表
		ToPortAPIs(*bizmodel.Strategy, []*bizmodel.StrategyPortLevel) []*api.DomainStrategyItem
		// ToHTTPLevelAPI 转换为API对象
		ToHTTPLevelAPI(*bizmodel.Strategy, *bizmodel.StrategyHTTPLevel) *api.HttpStrategyItem
		// ToHTTPLevelAPIs 转换为API对象列表
		ToHTTPLevelAPIs(*bizmodel.Strategy, []*bizmodel.StrategyHTTPLevel) []*api.HttpStrategyItem
		// ToPingAPI 转换为API对象
		ToPingAPI(*bizmodel.Strategy, *bizmodel.StrategyPingLevel) *api.PingStrategyItem
		// ToPingAPIs 转换为API对象列表
		ToPingAPIs(*bizmodel.Strategy, []*bizmodel.StrategyPingLevel) []*api.PingStrategyItem

		// ToLogAPI 转换为API对象
		ToLogAPI(*bizmodel.Strategy, *bizmodel.StrategyLogsLevel) *api.LogsStrategyItem
		// ToLogAPIs 转换为API对象列表
		ToLogAPIs(*bizmodel.Strategy, []*bizmodel.StrategyLogsLevel) []*api.LogsStrategyItem
	}

	doStrategyLevelsBuilder struct {
		ctx context.Context
	}

	// IBoStrategyBuilder 策略业务对象构造器
	IBoStrategyBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(...*bo.Strategy) *houyistrategyapi.PushStrategyRequest
	}

	boStrategyBuilder struct {
		ctx context.Context
	}
)

func (d *doStrategyLevelsBuilder) ToLogAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyLogsLevel) []*api.LogsStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyLogsLevel) *api.LogsStrategyItem {
		return d.ToLogAPI(strategy, level)
	})
}

func (d *doStrategyLevelsBuilder) ToLogAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyLogsLevel) *api.LogsStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.GetAlarmNoticeGroups())+len(level.GetAlarmGroupList()))
	for _, group := range strategy.GetAlarmNoticeGroups() {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.GetAlarmGroupList() {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)

	return &api.LogsStrategyItem{
		StrategyType:     api.StrategyType(strategy.StrategyType),
		StrategyID:       strategy.ID,
		TeamID:           strategy.TeamID,
		Status:           api.Status(strategy.Status),
		Alert:            strategy.Name,
		LevelId:          level.GetLevel().ID,
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		Expr:             strategy.Expr,
		For:              durationpb.New(time.Duration(level.Duration) * time.Second),
		Datasource:       NewParamsBuild(d.ctx).DatasourceModuleBuilder().DatasourceBuilder().ToAPIs(strategy.Datasource),
		ReceiverGroupIDs: receiverGroupIDs,
	}
}

func (d *doStrategyLevelsBuilder) ToPingAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyPingLevel) *api.PingStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.GetAlarmNoticeGroups())+len(level.GetAlarmGroupList()))
	for _, group := range strategy.GetAlarmNoticeGroups() {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.GetAlarmGroupList() {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)
	return &api.PingStrategyItem{
		StrategyType:     api.StrategyType(strategy.StrategyType),
		StrategyID:       strategy.ID,
		TeamID:           strategy.TeamID,
		Status:           api.Status(strategy.Status),
		Alert:            strategy.Name,
		LevelId:          level.GetLevel().ID,
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		ReceiverGroupIDs: receiverGroupIDs,
		Address:          strategy.Expr,
		TotalCount:       int64(level.Total),
		SuccessCount:     int64(level.Success),
		LossRate:         level.LossRate,
		AvgDelay:         int64(level.AvgDelay),
		MaxDelay:         int64(level.MaxDelay),
		MinDelay:         int64(level.MinDelay),
		StdDev:           int64(level.StdDev),
	}
}

func (d *doStrategyLevelsBuilder) ToPingAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyPingLevel) []*api.PingStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyPingLevel) *api.PingStrategyItem {
		return d.ToPingAPI(strategy, level)
	})
}

func (d *doStrategyLevelsBuilder) ToMetricAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyMetricLevel) *api.MetricStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.GetAlarmNoticeGroups())+len(level.GetAlarmGroupList()))
	for _, group := range strategy.AlarmNoticeGroups {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.AlarmGroupList {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)
	return &api.MetricStrategyItem{
		Alert:            strategy.Name,
		Expr:             strategy.Expr,
		For:              durationpb.New(time.Duration(level.Duration) * time.Second),
		Count:            level.Count,
		SustainType:      api.SustainType(level.SustainType),
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		Status:           api.Status(strategy.Status),
		Datasource:       NewParamsBuild(d.ctx).DatasourceModuleBuilder().DatasourceBuilder().ToAPIs(strategy.Datasource),
		Condition:        api.Condition(level.Condition),
		Threshold:        level.Threshold,
		LevelId:          level.GetLevel().ID,
		TeamID:           strategy.TeamID,
		ReceiverGroupIDs: receiverGroupIDs,
		LabelNotices:     NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().AlarmNoticeGroupItemBuilder().ToAPIs(level.GetLabelNoticeList()),
		StrategyType:     api.StrategyType(strategy.StrategyType),
		StrategyID:       strategy.ID,
	}
}

func (d *doStrategyLevelsBuilder) ToMetricAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyMetricLevel) []*api.MetricStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyMetricLevel) *api.MetricStrategyItem {
		return d.ToMetricAPI(strategy, level)
	})
}

func (d *doStrategyLevelsBuilder) ToEventAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyEventLevel) *api.EventStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.AlarmNoticeGroups)+len(level.AlarmGroupList))
	for _, group := range strategy.AlarmNoticeGroups {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.AlarmGroupList {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)
	return &api.EventStrategyItem{
		StrategyType:     api.StrategyType(strategy.StrategyType),
		StrategyID:       strategy.ID,
		TeamID:           strategy.TeamID,
		Status:           api.Status(strategy.Status),
		Alert:            strategy.Name,
		LevelId:          level.GetLevel().ID,
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		ReceiverGroupIDs: receiverGroupIDs,
		Value:            level.Value,
		Condition:        api.EventCondition(level.Condition),
		DataType:         api.EventDataType(level.DataType),
		Topic:            strategy.Expr,
		Datasource:       NewParamsBuild(d.ctx).DatasourceModuleBuilder().DatasourceBuilder().ToAPIs(strategy.Datasource),
		DataKey:          level.PathKey,
	}
}

func (d *doStrategyLevelsBuilder) ToEventAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyEventLevel) []*api.EventStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyEventLevel) *api.EventStrategyItem {
		return d.ToEventAPI(strategy, level)
	})
}

func (d *doStrategyLevelsBuilder) ToDomainAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyDomainLevel) *api.DomainStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.AlarmNoticeGroups)+len(level.AlarmGroupList))
	for _, group := range strategy.AlarmNoticeGroups {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.AlarmGroupList {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)
	return &api.DomainStrategyItem{
		StrategyID:       strategy.ID,
		LevelId:          level.GetLevel().ID,
		TeamID:           strategy.TeamID,
		ReceiverGroupIDs: receiverGroupIDs,
		Status:           api.Status(strategy.Status),
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		Threshold:        level.Threshold,
		Domain:           strategy.Expr,
		Condition:        api.Condition(level.Condition),
		Alert:            strategy.Name,
		Port:             443,
		StrategyType:     api.StrategyType(strategy.StrategyType),
	}
}

func (d *doStrategyLevelsBuilder) ToDomainAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyDomainLevel) []*api.DomainStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyDomainLevel) *api.DomainStrategyItem {
		return d.ToDomainAPI(strategy, level)
	})
}

func (d *doStrategyLevelsBuilder) ToPortAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyPortLevel) *api.DomainStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.AlarmNoticeGroups)+len(level.AlarmGroupList))
	for _, group := range strategy.AlarmNoticeGroups {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.AlarmGroupList {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)
	return &api.DomainStrategyItem{
		StrategyID:       strategy.ID,
		LevelId:          level.GetLevel().GetID(),
		TeamID:           strategy.TeamID,
		ReceiverGroupIDs: receiverGroupIDs,
		Status:           api.Status(strategy.Status),
		Labels:           strategy.Labels.Map(),
		Annotations:      strategy.Annotations.Map(),
		Threshold:        level.Threshold,
		Domain:           strategy.Expr,
		Condition:        api.Condition(vobj.ConditionEQ),
		Alert:            strategy.Name,
		Port:             level.Port,
		StrategyType:     api.StrategyType(strategy.StrategyType),
	}
}

func (d *doStrategyLevelsBuilder) ToPortAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyPortLevel) []*api.DomainStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyPortLevel) *api.DomainStrategyItem {
		return d.ToPortAPI(strategy, level)
	})
}

func (d *doStrategyLevelsBuilder) ToHTTPLevelAPI(strategy *bizmodel.Strategy, level *bizmodel.StrategyHTTPLevel) *api.HttpStrategyItem {
	if types.IsNil(strategy) || types.IsNil(level) || types.IsNil(d) {
		return nil
	}
	receiverGroupIDs := make([]uint32, 0, len(strategy.AlarmNoticeGroups)+len(level.AlarmGroupList))
	for _, group := range strategy.GetAlarmNoticeGroups() {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	for _, group := range level.GetAlarmGroupList() {
		receiverGroupIDs = append(receiverGroupIDs, group.ID)
	}
	receiverGroupIDs = types.SliceUnique(receiverGroupIDs)
	headers := make(map[string]string, len(level.Headers))
	for _, header := range level.Headers {
		headers[header.Name] = header.Value
	}
	return &api.HttpStrategyItem{
		StrategyType:          api.StrategyType(strategy.StrategyType),
		Url:                   strategy.Expr,
		StrategyID:            strategy.ID,
		LevelId:               level.GetLevel().GetID(),
		TeamID:                strategy.TeamID,
		ReceiverGroupIDs:      receiverGroupIDs,
		Status:                api.Status(strategy.Status),
		Labels:                strategy.Labels.Map(),
		Annotations:           strategy.Annotations.Map(),
		StatusCode:            level.StatusCode,
		Alert:                 strategy.Name,
		Headers:               headers,
		Body:                  level.Body,
		Method:                level.Method.String(),
		ResponseTime:          level.ResponseTime,
		StatusCodeCondition:   api.Condition(level.StatusCodeCondition),
		ResponseTimeCondition: api.Condition(level.ResponseTimeCondition),
	}
}

func (d *doStrategyLevelsBuilder) ToHTTPLevelAPIs(strategy *bizmodel.Strategy, levels []*bizmodel.StrategyHTTPLevel) []*api.HttpStrategyItem {
	return types.SliceTo(levels, func(level *bizmodel.StrategyHTTPLevel) *api.HttpStrategyItem {
		return d.ToHTTPLevelAPI(strategy, level)
	})
}

func (s *strategyModuleBuilder) DoStrategyLevelsBuilder() IDoStrategyLevelsBuilder {
	return &doStrategyLevelsBuilder{ctx: s.ctx}
}

func (b *boStrategyBuilder) ToAPI(strategies ...*bo.Strategy) *houyistrategyapi.PushStrategyRequest {
	metricLevels := make([]*api.MetricStrategyItem, 0, len(strategies))
	domainLevels := make([]*api.DomainStrategyItem, 0, len(strategies))
	httpLevels := make([]*api.HttpStrategyItem, 0, len(strategies))
	pingLevels := make([]*api.PingStrategyItem, 0, len(strategies))
	eventLevels := make([]*api.EventStrategyItem, 0, len(strategies))
	logsLevels := make([]*api.LogsStrategyItem, 0, len(strategies))

	for _, strategy := range strategies {
		if types.IsNotNil(strategy.MetricLevel) {
			metricLevels = append(metricLevels, strategy.MetricLevel)
		}
		if types.IsNotNil(strategy.DomainLevel) {
			domainLevels = append(domainLevels, strategy.DomainLevel)
		}
		if types.IsNotNil(strategy.HTTPLevel) {
			httpLevels = append(httpLevels, strategy.HTTPLevel)
		}
		if types.IsNotNil(strategy.PingLevel) {
			pingLevels = append(pingLevels, strategy.PingLevel)
		}
		if types.IsNotNil(strategy.EventLevel) {
			eventLevels = append(eventLevels, strategy.EventLevel)
		}
		if types.IsNotNil(strategy.PortLevel) {
			domainLevels = append(domainLevels, strategy.PortLevel)
		}
		if types.IsNotNil(strategy.LogsLevel) {
			logsLevels = append(logsLevels, strategy.LogsLevel)
		}
	}

	return &houyistrategyapi.PushStrategyRequest{
		MetricStrategies: metricLevels,
		DomainStrategies: domainLevels,
		HttpStrategies:   httpLevels,
		PingStrategies:   pingLevels,
		EventStrategies:  eventLevels,
		LogStrategies:    logsLevels,
	}
}

func (d *doStrategyLevelBuilder) ToDomainAPI(domain *bizmodel.StrategyDomainLevel) *adminapi.StrategyDomainLevelItem {
	if types.IsNil(domain) {
		return nil
	}
	return &adminapi.StrategyDomainLevelItem{
		Level:        NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(domain.Level),
		AlarmPages:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(domain.AlarmPageList, func(item *bizmodel.SysDict) imodel.IDict { return item })),
		AlarmGroups:  NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(domain.AlarmGroupList),
		LabelNotices: nil,
		Threshold:    domain.Threshold,
		Condition:    api.Condition(domain.Condition),
	}
}

func (d *doStrategyLevelBuilder) ToDomainAPIs(domains []*bizmodel.StrategyDomainLevel) []*adminapi.StrategyDomainLevelItem {
	return types.SliceTo(domains, func(domain *bizmodel.StrategyDomainLevel) *adminapi.StrategyDomainLevelItem {
		return d.ToDomainAPI(domain)
	})
}

func (d *doStrategyLevelBuilder) ToPortAPI(port *bizmodel.StrategyPortLevel) *adminapi.StrategyPortLevelItem {
	if types.IsNil(port) {
		return nil
	}

	return &adminapi.StrategyPortLevelItem{
		Level: NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(port.Level),
		AlarmPages: NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(port.AlarmPageList, func(item *bizmodel.SysDict) imodel.IDict {
			return item
		})),
		AlarmGroups:  NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(port.AlarmGroupList),
		LabelNotices: nil,
		Threshold:    port.Threshold,
		Port:         port.Port,
	}
}

func (d *doStrategyLevelBuilder) ToPortAPIs(ports []*bizmodel.StrategyPortLevel) []*adminapi.StrategyPortLevelItem {
	return types.SliceTo(ports, func(port *bizmodel.StrategyPortLevel) *adminapi.StrategyPortLevelItem {
		return d.ToPortAPI(port)
	})
}

func (d *doStrategyLevelBuilder) ToHTTPLevelAPI(http *bizmodel.StrategyHTTPLevel) *adminapi.StrategyHTTPLevelItem {
	if types.IsNil(http) {
		return nil
	}
	header := make(map[string]string, len(http.Headers))
	for _, item := range http.Headers {
		header[item.Name] = item.Value
	}
	return &adminapi.StrategyHTTPLevelItem{
		StatusCode:            http.StatusCode,
		Level:                 NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(http.Level),
		AlarmPages:            NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(http.AlarmPageList, func(item *bizmodel.SysDict) imodel.IDict { return item })),
		AlarmGroups:           NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(http.AlarmGroupList),
		ResponseTime:          http.ResponseTime,
		Headers:               header,
		Body:                  http.Body,
		QueryParams:           http.QueryParams,
		Method:                http.Method.String(),
		StatusCodeCondition:   api.Condition(http.StatusCodeCondition),
		ResponseTimeCondition: api.Condition(http.ResponseTimeCondition),
	}
}

func (d *doStrategyLevelBuilder) ToHTTPLevelAPIs(https []*bizmodel.StrategyHTTPLevel) []*adminapi.StrategyHTTPLevelItem {
	return types.SliceTo(https, func(http *bizmodel.StrategyHTTPLevel) *adminapi.StrategyHTTPLevelItem {
		return d.ToHTTPLevelAPI(http)
	})
}

func (m *mutationStrategyLevelBuilder) ToDomainBo(request *strategyapi.CreateStrategyDomainLevelRequest) *bo.CreateStrategyDomainLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return &bo.CreateStrategyDomainLevel{
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

func (b *boStrategyBuilder) ToEventAPI(strategy *bo.Strategy) *api.EventStrategyItem {
	if types.IsNil(strategy) || types.IsNil(strategy.EventLevel) || types.IsNil(b) {
		return nil
	}
	return strategy.EventLevel
}

func (d *doStrategyLevelBuilder) ToEventAPI(level *bizmodel.StrategyEventLevel) *adminapi.StrategyEventLevelItem {
	if types.IsNil(level) || types.IsNil(d) {
		return nil
	}

	return &adminapi.StrategyEventLevelItem{
		Threshold:    level.Value,
		Condition:    api.EventCondition(level.Condition),
		DataType:     api.EventDataType(level.DataType),
		Level:        NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(level.GetLevel()),
		AlarmPages:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(level.AlarmPageList, func(item *bizmodel.SysDict) imodel.IDict { return item })),
		AlarmGroups:  NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(level.AlarmGroupList),
		PathKey:      level.PathKey,
		LabelNotices: nil,
	}
}

func (d *doStrategyLevelBuilder) ToEventAPIs(levels []*bizmodel.StrategyEventLevel) []*adminapi.StrategyEventLevelItem {
	if types.IsNil(d) || types.IsNil(levels) {
		return nil
	}
	return types.SliceTo(levels, func(level *bizmodel.StrategyEventLevel) *adminapi.StrategyEventLevelItem {
		return d.ToEventAPI(level)
	})
}

func (m *mutationStrategyLevelBuilder) ToEventBo(request *strategyapi.CreateStrategyEventLevelRequest) *bo.CreateStrategyEventLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return &bo.CreateStrategyEventLevel{
		Value:         request.GetThreshold(),
		Condition:     vobj.EventCondition(request.GetCondition()),
		EventDataType: vobj.EventDataType(request.DataType),
		LevelID:       request.GetLevelId(),
		AlarmPageIds:  request.GetAlarmPageIds(),
		AlarmGroupIds: request.GetAlarmGroupIds(),
		StrategyID:    m.StrategyID,
		PathKey:       request.GetPathKey(),
	}
}

func (m *mutationStrategyLevelBuilder) ToEventBos(request []*strategyapi.CreateStrategyEventLevelRequest) []*bo.CreateStrategyEventLevel {
	if types.IsNil(request) || types.IsNil(m) {
		return nil
	}
	return types.SliceTo(request, func(item *strategyapi.CreateStrategyEventLevelRequest) *bo.CreateStrategyEventLevel {
		return m.ToEventBo(item)
	})
}

func (d *doStrategyBuilder) ToBos(strategy *bizmodel.Strategy) []*bo.Strategy {
	switch strategy.StrategyType {
	case vobj.StrategyTypeMetric:
		return types.SliceTo(strategy.GetLevel().GetStrategyMetricsLevelList(), func(item *bizmodel.StrategyMetricLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				MetricLevel:  NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToMetricAPI(strategy, item),
			}
		})
	case vobj.StrategyTypeHTTP:
		return types.SliceTo(strategy.GetLevel().GetStrategyHTTPLevelList(), func(item *bizmodel.StrategyHTTPLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				HTTPLevel:    NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToHTTPLevelAPI(strategy, item),
			}
		})
	case vobj.StrategyTypeEvent:
		return types.SliceTo(strategy.GetLevel().GetStrategyEventLevelList(), func(item *bizmodel.StrategyEventLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				EventLevel:   NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToEventAPI(strategy, item),
			}
		})
	case vobj.StrategyTypeDomainPort:
		return types.SliceTo(strategy.GetLevel().GetStrategyPortLevelList(), func(item *bizmodel.StrategyPortLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				PortLevel:    NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToPortAPI(strategy, item),
			}
		})
	case vobj.StrategyTypeDomainCertificate:
		return types.SliceTo(strategy.GetLevel().GetStrategyDomainLevelList(), func(item *bizmodel.StrategyDomainLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				DomainLevel:  NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToDomainAPI(strategy, item),
			}
		})
	case vobj.StrategyTypePing:
		return types.SliceTo(strategy.GetLevel().GetStrategyPingLevelList(), func(item *bizmodel.StrategyPingLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				PingLevel:    NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToPingAPI(strategy, item),
			}
		})
	case vobj.StrategyTypeLogs:
		return types.SliceTo(strategy.GetLevel().GetStrategyLogLevelList(), func(item *bizmodel.StrategyLogsLevel) *bo.Strategy {
			return &bo.Strategy{
				TeamID:       strategy.TeamID,
				StrategyID:   strategy.ID,
				StrategyType: strategy.StrategyType,
				LogsLevel:    NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelsBuilder().ToLogAPI(strategy, item),
			}
		})
	default:
		return nil
	}
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
		StrategyTemplateID: m.StrategyID,
		Duration:           request.GetDuration(),
		Count:              request.Count,
		SustainType:        vobj.Sustain(request.SustainType),
		Condition:          vobj.Condition(request.Condition),
		Threshold:          request.Threshold,
		LevelID:            request.LevelId,
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

func (d *doStrategyLevelBuilder) ToMetricAPI(level *bizmodel.StrategyMetricLevel) *adminapi.StrategyMetricLevelItem {
	if types.IsNil(d) || types.IsNil(level) {
		return nil
	}

	return &adminapi.StrategyMetricLevelItem{
		Duration:     level.Duration,
		Count:        level.Count,
		SustainType:  api.SustainType(level.SustainType),
		Level:        NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelect(level.GetLevel()),
		AlarmPages:   NewParamsBuild(d.ctx).DictModuleBuilder().DoDictBuilder().ToSelects(types.SliceTo(level.GetAlarmPageList(), func(item *bizmodel.SysDict) imodel.IDict { return item })),
		Threshold:    level.Threshold,
		Condition:    api.Condition(level.Condition),
		AlarmGroups:  NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(level.GetAlarmGroupList()),
		LabelNotices: NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoLabelNoticeBuilder().ToAPIs(level.GetLabelNoticeList()),
	}
}

func (d *doStrategyLevelBuilder) ToMetricAPIs(levels []*bizmodel.StrategyMetricLevel) []*adminapi.StrategyMetricLevelItem {
	if types.IsNil(d) || types.IsNil(levels) {
		return nil
	}
	return types.SliceTo(levels, func(level *bizmodel.StrategyMetricLevel) *adminapi.StrategyMetricLevelItem {
		return d.ToMetricAPI(level)
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

func (d *doStrategyLevelTemplateBuilder) ToAPI(template *model.StrategyLevelTemplate) *adminapi.StrategyLevelTemplateItem {
	if types.IsNil(d) || types.IsNil(template) {
		return nil
	}

	userMap := getUsers(d.ctx, template.CreatorID)

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

	return types.SliceTo(templates, func(item *model.StrategyLevelTemplate) *adminapi.StrategyLevelTemplateItem {
		return d.ToAPI(item)
	})
}

func (s *strategyModuleBuilder) DoStrategyLevelTemplateBuilder() IDoStrategyLevelTemplateBuilder {
	return &doStrategyLevelTemplateBuilder{ctx: s.ctx}
}

func (d *doTemplateStrategyBuilder) ToAPI(template *model.StrategyTemplate) *adminapi.StrategyTemplateItem {
	if types.IsNil(d) || types.IsNil(template) {
		return nil
	}

	userMap := getUsers(d.ctx, template.CreatorID)
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

	return types.SliceTo(templates, func(item *model.StrategyTemplate) *adminapi.StrategyTemplateItem {
		return d.ToAPI(item)
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
			Labels:                 label.NewLabels(u.GetLabels()),
			Annotations:            label.NewAnnotations(u.GetAnnotations()),
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
		Labels:                 label.NewLabels(c.GetLabels()),
		Annotations:            label.NewAnnotations(c.GetAnnotations()),
		StrategyLevelTemplates: NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelTemplateItems().ToBos(c.GetLevels()),
		CategoriesIDs:          c.GetCategoriesIds(),
	}
}

func (d *doStrategyBuilder) ToAPI(strategy *bizmodel.Strategy) *adminapi.StrategyItem {
	if types.IsNil(d) || types.IsNil(strategy) {
		return nil
	}

	userMap := getUsers(d.ctx, strategy.CreatorID)
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
		MetricLevels:      NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToMetricAPIs(strategy.GetLevel().GetStrategyMetricsLevelList()),
		EventLevels:       NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToEventAPIs(strategy.GetLevel().GetStrategyEventLevelList()),
		PortLevels:        NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToPortAPIs(strategy.GetLevel().GetStrategyPortLevelList()),
		HttpLevels:        NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToHTTPLevelAPIs(strategy.GetLevel().GetStrategyHTTPLevelList()),
		DomainLevels:      NewParamsBuild(d.ctx).StrategyModuleBuilder().DoStrategyLevelBuilder().ToDomainAPIs(strategy.GetLevel().GetStrategyDomainLevelList()),
	}

	return strategyItem
}

func (d *doStrategyBuilder) ToAPIs(strategies []*bizmodel.Strategy) []*adminapi.StrategyItem {
	if types.IsNil(d) || types.IsNil(strategies) {
		return nil
	}

	return types.SliceTo(strategies, func(item *bizmodel.Strategy) *adminapi.StrategyItem {
		return d.ToAPI(item)
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
		Labels:         label.NewLabels(c.GetLabels()),
		Annotations:    label.NewAnnotations(c.GetAnnotations()),
		Expr:           c.GetExpr(),
		CategoriesIds:  c.GetCategoriesIds(),
		AlarmGroupIds:  c.GetAlarmGroupIds(),
		StrategyType:   vobj.StrategyType(c.GetStrategyType()),
		MetricLevels:   NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToMetricBos(c.GetStrategyMetricLevels()),
		EventLevels:    NewParamsBuild(c.ctx).StrategyModuleBuilder().APIMutationStrategyLevelItems().ToEventBos(c.GetStrategyEventLevels()),
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

func (d *doStrategyGroupBuilder) ToAPI(group *bizmodel.StrategyGroup) *adminapi.StrategyGroupItem {
	if types.IsNil(d) || types.IsNil(group) {
		return nil
	}
	userMap := getUsers(d.ctx, group.CreatorID)
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

	return types.SliceTo(groups, func(item *bizmodel.StrategyGroup) *adminapi.StrategyGroupItem {
		return d.ToAPI(item)
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
