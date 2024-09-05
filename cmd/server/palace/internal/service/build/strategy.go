package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// StrategyModelBuilder 策略模型构建器
	StrategyModelBuilder interface {
		ToAPI() *admin.StrategyItem
		ToBos() []*bo.Strategy
	}

	// BoStrategyBuilder bo 策略模型构建器
	BoStrategyBuilder interface {
		ToAPI() *api.Strategy
	}

	// BoStrategiesBuilder bo 策略模型构建器
	BoStrategiesBuilder interface {
		ToAPIs() []*api.Strategy
	}

	// BoStrategyModelBuilder bo 策略模型构建器
	BoStrategyModelBuilder interface {
		WithBoStrategy(*bo.Strategy) BoStrategyBuilder
		WithBoStrategies([]*bo.Strategy) BoStrategiesBuilder
	}

	// StrategyRequestBuilder 策略请求构建器
	StrategyRequestBuilder interface {
		ToCreateStrategyBO() *bo.CreateStrategyParams
		ToUpdateStrategyBO() *bo.UpdateStrategyParams
	}

	strategyBuilder struct {
		// model
		Strategy      *bizmodel.Strategy
		StrategyLevel *bizmodel.StrategyLevel
		// request
		CreateStrategy *strategyapi.CreateStrategyRequest
		UpdateStrategy *strategyapi.UpdateStrategyRequest

		// context
		ctx context.Context
	}

	boStrategyModelBuilder struct {
		ctx context.Context
	}

	// StrategyGroupModelBuilder 策略组模型构建器
	StrategyGroupModelBuilder interface {
		ToAPI() *admin.StrategyGroupItem
	}

	// StrategyGroupRequestBuilder 策略组请求构建器
	StrategyGroupRequestBuilder interface {
		ToCreateStrategyGroupBO() *bo.CreateStrategyGroupParams

		ToUpdateStrategyGroupBO() *bo.UpdateStrategyGroupParams

		ToListStrategyGroupBO() *bo.QueryStrategyGroupListParams
	}

	// StrategyGroupModuleBuilder StrategyGroupBuilder 策略组构建器
	StrategyGroupModuleBuilder interface {
		WithDoStrategyCount(items *bo.StrategyCountMap) DosStrategyGroupBuilder
		WithDoStrategyGroupList(items []*bizmodel.StrategyGroup) DosStrategyGroupBuilder
		WithDoStrategyGroup(items *bizmodel.StrategyGroup) DosStrategyGroupBuilder
	}

	strategyGroupModuleBuilder struct {
		ctx context.Context
	}

	// DosStrategyGroupBuilder   do  alarm group builder
	DosStrategyGroupBuilder interface {
		WithStrategyCountMap(item *bo.StrategyCountMap) DosStrategyGroupBuilder
		ToAPIs() []*admin.StrategyGroupItem
		ToAPI() *admin.StrategyGroupItem
	}

	// dosStrategyGroupBuilder dos  alarm group builder
	dosStrategyGroupBuilder struct {
		StrategyGroups []*bizmodel.StrategyGroup
		StrategyGroup  *bizmodel.StrategyGroup
		// 策略开启总数
		StrategyCountMap map[uint32]*bo.StrategyCountModel
		// 策略总数
		StrategyEnableMap map[uint32]*bo.StrategyCountModel
		ctx               context.Context
	}

	strategyGroupBuilder struct {
		// Model
		StrategyGroup            *bizmodel.StrategyGroup
		StrategyGroups           []*bizmodel.StrategyGroup
		StrategyCountModel       []*bo.StrategyCountModel
		StrategyEnableCountModel []*bo.StrategyCountModel

		// request
		CreateStrategyGroupRequest *strategyapi.CreateStrategyGroupRequest
		UpdateStrategyGroupRequest *strategyapi.UpdateStrategyGroupRequest
		ListStrategyGroupRequest   *strategyapi.ListStrategyGroupRequest

		// context
		ctx context.Context
	}

	// StrategyLevelModuleBuilder 策略等级模型构建器
	StrategyLevelModuleBuilder interface {
		WithAPIStrategyLevel(*strategyapi.CreateStrategyLevelRequest) APIAddStrategyLevelParamsBuilder
		WithDoStrategyLevel(*bizmodel.StrategyLevel) DoStrategyLevelBuilder
	}

	strategyLevelModuleBuilder struct {
		ctx context.Context
	}

	// APIAddStrategyLevelParamsBuilder 策略等级请求参数构建器
	APIAddStrategyLevelParamsBuilder interface {
		ToBo() *bo.CreateStrategyLevel
	}

	apiAddStrategyLevelParamsBuilder struct {
		params *strategyapi.CreateStrategyLevelRequest

		// context
		ctx context.Context
	}

	// DoStrategyLevelBuilder 策略等级模型构建器
	DoStrategyLevelBuilder interface {
		ToAPI() *admin.StrategyLevel
	}

	doStrategyLevelBuilder struct {
		strategyLevel *bizmodel.StrategyLevel

		ctx context.Context
	}

	boStrategyBuilder struct {
		strategy *bo.Strategy
		ctx      context.Context
	}

	boStrategiesBuilder struct {
		strategies []*bo.Strategy
		ctx        context.Context
	}
)

func (b *strategyBuilder) ToBos() []*bo.Strategy {
	if types.IsNil(b) || types.IsNil(b.Strategy) || types.IsNil(b.Strategy.StrategyLevel) || len(b.Strategy.StrategyLevel) == 0 {
		return nil
	}
	strategy := b.Strategy
	datasource := types.SliceToWithFilter(strategy.Datasource, func(s *bizmodel.Datasource) (*bo.Datasource, bool) {
		item := newDatasourceModelBuilder(b.ctx, s).ToBo()
		return item, !types.IsNil(item)
	})
	list := make([]*bo.Strategy, 0, len(strategy.StrategyLevel))
	for _, strategyLevel := range strategy.StrategyLevel {
		status := vobj.StatusDisable
		if strategyLevel.Status.IsEnable() &&
			!types.IsNil(strategy.StrategyGroup) &&
			strategy.StrategyGroup.Status.IsEnable() &&
			strategy.StrategyGroup.DeletedAt == 0 &&
			strategy.Status.IsEnable() &&
			strategyLevel.DeletedAt == 0 &&
			strategy.DeletedAt == 0 {
			status = vobj.StatusEnable
		}
		item := &bo.Strategy{
			ID:                         strategy.ID,
			LevelID:                    strategyLevel.ID,
			Alert:                      b.Strategy.Name,
			Expr:                       b.Strategy.Expr,
			For:                        strategyLevel.Duration,
			Count:                      strategyLevel.Count,
			SustainType:                strategyLevel.SustainType,
			MultiDatasourceSustainType: 0,
			Labels:                     strategy.Labels,
			Annotations:                strategy.Annotations,
			Interval:                   strategyLevel.Interval,
			Datasource:                 datasource,
			Status:                     status,
			Step:                       strategy.Step,
			Condition:                  strategyLevel.Condition,
			Threshold:                  strategyLevel.Threshold,
		}
		list = append(list, item)
	}
	return list
}

func (b *boStrategiesBuilder) ToAPIs() []*api.Strategy {
	if types.IsNil(b) || types.IsNil(b.strategies) {
		return nil
	}
	return types.SliceToWithFilter(b.strategies, func(s *bo.Strategy) (*api.Strategy, bool) {
		item := newBoStrategyBuilder(b.ctx, s).ToAPI()
		return item, !types.IsNil(item)
	})
}

func (b *boStrategyBuilder) ToAPI() *api.Strategy {
	if types.IsNil(b) || types.IsNil(b.strategy) {
		return nil
	}
	strategy := b.strategy
	return &api.Strategy{
		Alert:                      strategy.Alert,
		Expr:                       strategy.Expr,
		For:                        strategy.For.GetDuration(),
		Count:                      strategy.Count,
		SustainType:                api.SustainType(strategy.SustainType),
		MultiDatasourceSustainType: api.MultiDatasourceSustainType(strategy.MultiDatasourceSustainType),
		Labels:                     strategy.Labels.Map(),
		Annotations:                strategy.Annotations,
		Interval:                   strategy.Interval.GetDuration(),
		Datasource:                 newBoDatasourceBuilder(b.ctx, strategy.Datasource).ToAPIs(),
		Id:                         strategy.ID,
		Status:                     api.Status(strategy.Status),
		Step:                       strategy.Step,
		Condition:                  api.Condition(strategy.Condition),
		Threshold:                  strategy.Threshold,
		LevelID:                    strategy.LevelID,
	}
}

func (b *boStrategyModelBuilder) WithBoStrategy(strategy *bo.Strategy) BoStrategyBuilder {
	return newBoStrategyBuilder(b.ctx, strategy)
}

func (b *boStrategyModelBuilder) WithBoStrategies(strategies []*bo.Strategy) BoStrategiesBuilder {
	return newBoStrategiesBuilder(b.ctx, strategies)
}

func (d doStrategyLevelBuilder) ToAPI() *admin.StrategyLevel {
	if types.IsNil(d) || types.IsNil(d.strategyLevel) {
		return nil
	}
	level := d.strategyLevel
	strategyLevel := &admin.StrategyLevel{
		Duration:    level.Duration.GetDuration(),
		Count:       level.Count,
		SustainType: api.SustainType(level.SustainType),
		Interval:    level.Interval.GetDuration(),
		Status:      api.Status(level.Status),
		Id:          level.ID,
		LevelId:     level.LevelID,
		Threshold:   level.Threshold,
		StrategyId:  level.StrategyID,
		Condition:   api.Condition(level.Condition),
		AlarmPages: types.SliceTo(level.AlarmPage, func(page *bizmodel.SysDict) *admin.SelectItem {
			return NewBuilder().WithDict(page).ToAPISelect()
		}),
		AlarmGroups: types.SliceTo(level.AlarmGroups, func(group *bizmodel.AlarmGroup) *admin.AlarmGroupItem {
			return NewBuilder().AlarmGroupModule().WithDoAlarmGroup(group).ToAPI()
		}),
	}
	return strategyLevel
}

// ToAPI 转换为API层数据
func (b *strategyBuilder) ToAPI() *admin.StrategyItem {
	if types.IsNil(b) || types.IsNil(b.Strategy) {
		return nil
	}
	strategyLevels := types.SliceToWithFilter(b.Strategy.StrategyLevel, func(level *bizmodel.StrategyLevel) (*admin.StrategyLevel, bool) {
		return NewBuilder().StrategyLevelModelBuilder().WithDoStrategyLevel(level).ToAPI(), true
	})

	labelsNotice := types.SliceTo(b.Strategy.StrategyNoticeLabels, func(label *bizmodel.StrategyLabels) *admin.StrategyLabelsItem {
		strategyItem := &admin.StrategyLabelsItem{
			Name:  label.Name,
			Value: label.Value,
			AlarmGroups: types.SliceTo(label.AlarmGroups, func(group *bizmodel.AlarmGroup) *admin.AlarmGroupItem {
				return NewBuilder().AlarmGroupModule().WithDoAlarmGroup(group).ToAPI()
			}),
		}
		return strategyItem
	})

	return &admin.StrategyItem{
		Name:        b.Strategy.Name,
		Id:          b.Strategy.ID,
		Expr:        b.Strategy.Expr,
		Labels:      b.Strategy.Labels.Map(),
		Annotations: b.Strategy.Annotations,
		Datasource: types.SliceTo(b.Strategy.Datasource, func(datasource *bizmodel.Datasource) *admin.DatasourceItem {
			return NewBuilder().WithContext(b.ctx).WithDoDatasource(datasource).ToAPI()
		}),
		StrategyTemplateId: b.Strategy.StrategyTemplateID,
		Levels:             strategyLevels,
		Status:             api.Status(b.Strategy.Status),
		Step:               b.Strategy.Step,
		SourceType:         api.TemplateSourceType(b.Strategy.StrategyTemplateSource),
		Categories: types.SliceTo(b.Strategy.Categories, func(dict *bizmodel.SysDict) *admin.Dict {
			return NewBuilder().WithContext(b.ctx).WithDict(dict).ToAPI()
		}),
		AlarmGroups: types.SliceTo(b.Strategy.AlarmGroups, func(alarmGroup *bizmodel.AlarmGroup) *admin.AlarmGroupItem {
			return NewBuilder().WithContext(b.ctx).AlarmGroupModule().WithDoAlarmGroup(alarmGroup).ToAPI()
		}),
		StrategyLabels: labelsNotice,
	}
}

func (b *strategyBuilder) ToCreateStrategyBO() *bo.CreateStrategyParams {
	if types.IsNil(b) || types.IsNil(b.CreateStrategy) {
		return nil
	}
	strategyLevels := make([]*bo.CreateStrategyLevel, 0, len(b.CreateStrategy.GetStrategyLevel()))
	for _, strategyLevel := range b.CreateStrategy.GetStrategyLevel() {
		strategyLevels = append(strategyLevels, &bo.CreateStrategyLevel{
			StrategyTemplateID: b.CreateStrategy.TemplateId,
			Count:              strategyLevel.GetCount(),
			Duration:           types.NewDuration(strategyLevel.GetDuration()),
			SustainType:        vobj.Sustain(strategyLevel.SustainType),
			Interval:           types.NewDuration(strategyLevel.GetInterval()),
			Condition:          vobj.Condition(strategyLevel.GetCondition()),
			Threshold:          strategyLevel.GetThreshold(),
			Status:             vobj.Status(strategyLevel.GetStatus()),
			LevelID:            strategyLevel.GetLevelId(),
			AlarmPageIds:       strategyLevel.GetAlarmPageIds(),
			AlarmGroupIds:      strategyLevel.GetAlarmGroupIds(),
		})
	}
	return &bo.CreateStrategyParams{
		TemplateID:    b.CreateStrategy.GetTemplateId(),
		GroupID:       b.CreateStrategy.GetGroupId(),
		Name:          b.CreateStrategy.GetName(),
		Remark:        b.CreateStrategy.GetRemark(),
		Status:        vobj.Status(b.CreateStrategy.GetStatus()),
		Step:          b.CreateStrategy.GetStep(),
		SourceType:    vobj.TemplateSourceType(b.CreateStrategy.GetSourceType()),
		DatasourceIDs: b.CreateStrategy.GetDatasourceIds(),
		Labels:        vobj.NewLabels(b.CreateStrategy.GetLabels()),
		Annotations:   b.CreateStrategy.GetAnnotations(),
		Expr:          b.CreateStrategy.GetExpr(),
		CategoriesIds: b.CreateStrategy.GetCategoriesIds(),
		AlarmGroupIds: b.CreateStrategy.GetAlarmGroupIds(),
		StrategyLevel: strategyLevels,
		StrategyLabels: types.SliceTo(b.CreateStrategy.GetStrategyLabels(), func(strategyLabel *strategyapi.CreateStrategyLabelsRequest) *bo.StrategyLabels {
			return &bo.StrategyLabels{
				Name:          strategyLabel.GetName(),
				Value:         strategyLabel.GetValue(),
				AlarmGroupIds: strategyLabel.GetAlarmGroupIds(),
			}
		}),
	}
}

func (s *apiAddStrategyLevelParamsBuilder) ToBo() *bo.CreateStrategyLevel {
	if types.IsNil(s) || types.IsNil(s.params) {
		return nil
	}
	createLevel := s.params
	return &bo.CreateStrategyLevel{
		Count:        createLevel.GetCount(),
		Duration:     types.NewDuration(createLevel.GetDuration()),
		SustainType:  vobj.Sustain(createLevel.SustainType),
		Interval:     types.NewDuration(createLevel.GetInterval()),
		Condition:    vobj.Condition(createLevel.GetCondition()),
		Threshold:    createLevel.GetThreshold(),
		Status:       vobj.Status(createLevel.GetStatus()),
		LevelID:      createLevel.GetLevelId(),
		AlarmPageIds: createLevel.GetAlarmPageIds(),
	}
}

func (b *strategyBuilder) ToUpdateStrategyBO() *bo.UpdateStrategyParams {
	if types.IsNil(b) || types.IsNil(b.UpdateStrategy) {
		return nil
	}
	strategyLevels := make([]*bo.CreateStrategyLevel, 0, len(b.UpdateStrategy.GetData().GetStrategyLevel()))
	for _, strategyLevel := range b.UpdateStrategy.GetData().GetStrategyLevel() {
		strategyLevels = append(strategyLevels, &bo.CreateStrategyLevel{
			StrategyTemplateID: b.UpdateStrategy.GetData().TemplateId,
			Count:              strategyLevel.GetCount(),
			Duration:           types.NewDuration(strategyLevel.GetDuration()),
			SustainType:        vobj.Sustain(strategyLevel.SustainType),
			Interval:           types.NewDuration(strategyLevel.GetInterval()),
			Condition:          vobj.Condition(strategyLevel.GetCondition()),
			Threshold:          strategyLevel.GetThreshold(),
			Status:             vobj.Status(strategyLevel.GetStatus()),
			LevelID:            strategyLevel.GetLevelId(),
			StrategyID:         b.UpdateStrategy.GetId(),
		})
	}
	return &bo.UpdateStrategyParams{
		ID: b.UpdateStrategy.GetId(),
		UpdateParam: bo.CreateStrategyParams{
			TemplateID:    b.UpdateStrategy.GetData().GetTemplateId(),
			GroupID:       b.UpdateStrategy.GetData().GetGroupId(),
			Name:          b.UpdateStrategy.GetData().GetName(),
			Remark:        b.UpdateStrategy.GetData().GetRemark(),
			Status:        vobj.Status(b.UpdateStrategy.GetData().GetStatus()),
			Step:          b.UpdateStrategy.GetData().GetStep(),
			SourceType:    vobj.TemplateSourceType(b.UpdateStrategy.GetData().GetSourceType()),
			DatasourceIDs: b.UpdateStrategy.GetData().GetDatasourceIds(),
			Labels:        vobj.NewLabels(b.UpdateStrategy.GetData().GetLabels()),
			Annotations:   b.UpdateStrategy.GetData().GetAnnotations(),
			Expr:          b.UpdateStrategy.GetData().GetExpr(),
			CategoriesIds: b.UpdateStrategy.GetData().GetCategoriesIds(),
			AlarmGroupIds: b.UpdateStrategy.GetData().GetAlarmGroupIds(),
			StrategyLevel: strategyLevels,
			StrategyLabels: types.SliceTo(b.UpdateStrategy.GetData().GetStrategyLabels(), func(strategyLabel *strategyapi.CreateStrategyLabelsRequest) *bo.StrategyLabels {
				return &bo.StrategyLabels{
					Name:          strategyLabel.GetName(),
					Value:         strategyLabel.GetValue(),
					AlarmGroupIds: strategyLabel.GetAlarmGroupIds(),
				}
			}),
		},
	}
}

func (b *strategyGroupBuilder) ToAPI() *admin.StrategyGroupItem {
	if types.IsNil(b) || types.IsNil(b.StrategyGroup) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	groupItem := &admin.StrategyGroupItem{
		Id:        b.StrategyGroup.ID,
		Name:      b.StrategyGroup.Name,
		Remark:    b.StrategyGroup.Remark,
		Status:    api.Status(b.StrategyGroup.Status),
		CreatedAt: b.StrategyGroup.CreatedAt.String(),
		UpdatedAt: b.StrategyGroup.UpdatedAt.String(),
		Creator:   NewBuilder().WithAPIUserBo(cache.GetUser(b.ctx, b.StrategyGroup.CreatorID)).GetUsername(),
	}
	return groupItem
}

func (b *dosStrategyGroupBuilder) ToAPI() *admin.StrategyGroupItem {
	if types.IsNil(b) || types.IsNil(b.StrategyGroup) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	id := b.StrategyGroup.ID
	groupItem := &admin.StrategyGroupItem{
		Id:        id,
		Name:      b.StrategyGroup.Name,
		Remark:    b.StrategyGroup.Remark,
		Status:    api.Status(b.StrategyGroup.Status),
		CreatedAt: b.StrategyGroup.CreatedAt.String(),
		UpdatedAt: b.StrategyGroup.UpdatedAt.String(),
		Categories: types.SliceTo(b.StrategyGroup.Categories, func(category *bizmodel.SysDict) *admin.Dict {
			return NewBuilder().WithDict(category).ToAPI()
		}),
		Creator: NewBuilder().WithAPIUserBo(cache.GetUser(b.ctx, b.StrategyGroup.CreatorID)).GetUsername(),
	}
	count := b.StrategyCountMap[id]
	enableCount := b.StrategyEnableMap[id]
	if !types.IsNil(count) {
		groupItem.StrategyCount = count.Total
	}
	if !types.IsNil(enableCount) {
		groupItem.EnableStrategyCount = enableCount.Total
	}
	return groupItem
}

func (b *dosStrategyGroupBuilder) ToAPIs() []*admin.StrategyGroupItem {
	if types.IsNil(b) || types.IsNil(b.StrategyGroups) {
		return nil
	}
	return types.SliceTo(b.StrategyGroups, func(item *bizmodel.StrategyGroup) *admin.StrategyGroupItem {
		strategyGroup := NewBuilder().
			WithContext(b.ctx).
			StrategyGroupModuleBuilder().
			WithDoStrategyGroup(item).
			WithStrategyCountMap(&bo.StrategyCountMap{
				StrategyCountMap:  b.StrategyCountMap,
				StrategyEnableMap: b.StrategyEnableMap,
			}).ToAPI()
		return strategyGroup
	})
}

func (b *dosStrategyGroupBuilder) WithStrategyCountMap(item *bo.StrategyCountMap) DosStrategyGroupBuilder {
	if types.IsNil(b) || types.IsNil(item) {
		return b
	}
	b.StrategyCountMap = item.StrategyCountMap
	b.StrategyEnableMap = item.StrategyEnableMap
	return b
}

func (s *strategyGroupModuleBuilder) WithDoStrategyGroupList(items []*bizmodel.StrategyGroup) DosStrategyGroupBuilder {
	return &dosStrategyGroupBuilder{ctx: s.ctx, StrategyGroups: items}
}

func (s *strategyGroupModuleBuilder) WithDoStrategyGroup(item *bizmodel.StrategyGroup) DosStrategyGroupBuilder {
	return &dosStrategyGroupBuilder{ctx: s.ctx, StrategyGroup: item}
}

func (s *strategyGroupModuleBuilder) WithDoStrategyCount(item *bo.StrategyCountMap) DosStrategyGroupBuilder {
	return &dosStrategyGroupBuilder{ctx: s.ctx, StrategyCountMap: item.StrategyCountMap, StrategyEnableMap: item.StrategyEnableMap}
}

func (b *strategyGroupBuilder) ToCreateStrategyGroupBO() *bo.CreateStrategyGroupParams {
	if types.IsNil(b) || types.IsNil(b.CreateStrategyGroupRequest) {
		return nil
	}
	return &bo.CreateStrategyGroupParams{
		Name:          b.CreateStrategyGroupRequest.GetName(),
		Remark:        b.CreateStrategyGroupRequest.GetRemark(),
		Status:        b.CreateStrategyGroupRequest.GetStatus(),
		CategoriesIds: b.CreateStrategyGroupRequest.GetCategoriesIds(),
	}
}

func (b *strategyGroupBuilder) ToUpdateStrategyGroupBO() *bo.UpdateStrategyGroupParams {
	if types.IsNil(b) || types.IsNil(b.UpdateStrategyGroupRequest) {
		return nil
	}
	return &bo.UpdateStrategyGroupParams{
		ID: b.UpdateStrategyGroupRequest.GetId(),
		UpdateParam: bo.CreateStrategyGroupParams{
			Name:          b.UpdateStrategyGroupRequest.Update.GetName(),
			Remark:        b.UpdateStrategyGroupRequest.Update.GetRemark(),
			CategoriesIds: b.UpdateStrategyGroupRequest.Update.GetCategoriesIds(),
		},
	}
}

func (b *strategyGroupBuilder) ToListStrategyGroupBO() *bo.QueryStrategyGroupListParams {
	if types.IsNil(b) || types.IsNil(b.ListStrategyGroupRequest) {
		return nil
	}
	return &bo.QueryStrategyGroupListParams{
		Keyword:       b.ListStrategyGroupRequest.GetKeyword(),
		Status:        vobj.Status(b.ListStrategyGroupRequest.GetStatus()),
		Page:          types.NewPagination(b.ListStrategyGroupRequest.GetPagination()),
		CategoriesIds: b.ListStrategyGroupRequest.GetCategoriesIds(),
	}
}

func (s *strategyGroupModuleBuilder) WithDosStrategyGroup(item []*bizmodel.StrategyGroup) DosStrategyGroupBuilder {
	return &dosStrategyGroupBuilder{ctx: s.ctx, StrategyGroups: item}
}

// NewStrategyGroupModuleBuilder 创建策略组模块构建器
func NewStrategyGroupModuleBuilder(ctx context.Context) StrategyGroupModuleBuilder {
	return &strategyGroupModuleBuilder{ctx: ctx}
}

// NewStrategyLevelModelBuilder 创建策略等级构建器
func NewStrategyLevelModelBuilder(ctx context.Context) StrategyLevelModuleBuilder {
	return &strategyLevelModuleBuilder{ctx: ctx}
}

func (d *strategyLevelModuleBuilder) WithAPIStrategyLevel(request *strategyapi.CreateStrategyLevelRequest) APIAddStrategyLevelParamsBuilder {
	return &apiAddStrategyLevelParamsBuilder{params: request, ctx: d.ctx}
}

func (d *strategyLevelModuleBuilder) WithDoStrategyLevel(model *bizmodel.StrategyLevel) DoStrategyLevelBuilder {
	return &doStrategyLevelBuilder{strategyLevel: model, ctx: d.ctx}
}

func newBoStrategyModelBuilder(ctx context.Context) BoStrategyModelBuilder {
	return &boStrategyModelBuilder{ctx: ctx}
}

func newBoStrategiesBuilder(ctx context.Context, strategies []*bo.Strategy) BoStrategiesBuilder {
	return &boStrategiesBuilder{ctx: ctx, strategies: strategies}
}

func newBoStrategyBuilder(ctx context.Context, strategy *bo.Strategy) BoStrategyBuilder {
	return &boStrategyBuilder{ctx: ctx, strategy: strategy}
}
