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
		ToAPI(ctx context.Context) *admin.StrategyItem
	}

	// StrategyRequestBuilder 策略请求构建器
	StrategyRequestBuilder interface {
		ToCreateStrategyBO() *bo.CreateStrategyParams
		ToUpdateStrategyBO() *bo.UpdateStrategyParams
	}

	// StrategyLevelModelBuilder 策略等级模型构建器
	StrategyLevelModelBuilder interface {
		ToAPI() *admin.StrategyLevel
	}

	strategyLevelBuilder struct {
		// model
		*bizmodel.StrategyLevel
		ctx context.Context
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

	strategyGroupBuilder struct {
		// Model
		StrategyGroup *bizmodel.StrategyGroup

		// request
		CreateStrategyGroupRequest *strategyapi.CreateStrategyGroupRequest
		UpdateStrategyGroupRequest *strategyapi.UpdateStrategyGroupRequest
		ListStrategyGroupRequest   *strategyapi.ListStrategyGroupRequest

		// context
		ctx context.Context
	}
)

// ToAPI 转换为API层数据
func (b *strategyBuilder) ToAPI(ctx context.Context) *admin.StrategyItem {
	if types.IsNil(b) || types.IsNil(b.Strategy) {
		return nil
	}
	strategyLevels := types.SliceToWithFilter(b.Strategy.StrategyLevel, func(level *bizmodel.StrategyLevel) (*admin.StrategyLevel, bool) {
		return NewBuilder().WithAPIStrategyLevel(level).ToAPI(), true
	})

	return &admin.StrategyItem{
		Name:        b.Strategy.Name,
		Id:          b.Strategy.ID,
		Expr:        b.Strategy.Expr,
		Labels:      b.Strategy.Labels.Map(),
		Annotations: b.Strategy.Annotations,
		Datasource: types.SliceTo(b.Strategy.Datasource, func(datasource *bizmodel.Datasource) *admin.DatasourceItem {
			return NewBuilder().WithContext(ctx).WithDoDatasource(datasource).ToAPI()
		}),
		StrategyTemplateId: b.Strategy.StrategyTemplateID,
		Levels:             strategyLevels,
		Status:             api.Status(b.Strategy.Status),
		Step:               b.Strategy.Step,
		SourceType:         api.TemplateSourceType(b.Strategy.StrategyTemplateSource),
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
		})
	}
	return &bo.CreateStrategyParams{
		TeamID:        b.CreateStrategy.GetTeamId(),
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
		StrategyLevel: strategyLevels,
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
		})
	}
	return &bo.UpdateStrategyParams{
		TeamID: b.UpdateStrategy.GetData().GetTeamId(),
		ID:     b.UpdateStrategy.GetId(),
		UpdateParam: bo.CreateStrategyParams{
			TemplateID:    b.UpdateStrategy.GetData().GetTemplateId(),
			GroupID:       b.UpdateStrategy.GetData().GetGroupId(),
			Name:          b.UpdateStrategy.GetData().GetName(),
			Remark:        b.UpdateStrategy.GetData().GetRemark(),
			Status:        vobj.Status(b.UpdateStrategy.GetData().GetStatus()),
			Step:          b.UpdateStrategy.GetData().GetStep(),
			SourceType:    vobj.TemplateSourceType(b.UpdateStrategy.GetData().GetSourceType()),
			DatasourceIDs: b.UpdateStrategy.GetData().GetDatasourceIds(),
			StrategyLevel: strategyLevels,
		},
	}
}

func (b *strategyLevelBuilder) ToAPI() *admin.StrategyLevel {
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

func (b *strategyGroupBuilder) ToAPI() *admin.StrategyGroupItem {
	if types.IsNil(b) || types.IsNil(b.StrategyGroup) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	return &admin.StrategyGroupItem{
		Id:        b.StrategyGroup.ID,
		Name:      b.StrategyGroup.Name,
		Remark:    b.StrategyGroup.Remark,
		Status:    api.Status(b.StrategyGroup.Status),
		CreatedAt: b.StrategyGroup.CreatedAt.String(),
		UpdatedAt: b.StrategyGroup.UpdatedAt.String(),
		Creator:   NewBuilder().WithAPIUserBo(cache.GetUser(b.ctx, b.StrategyGroup.CreatorID)).ToAPI(),
	}
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
		TeamID:        b.CreateStrategyGroupRequest.GetTeamId(),
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
			TeamID:        b.UpdateStrategyGroupRequest.Update.GetTeamId(),
		},
		TeamID: b.UpdateStrategyGroupRequest.GetTeamId(),
	}
}

func (b *strategyGroupBuilder) ToListStrategyGroupBO() *bo.QueryStrategyGroupListParams {
	if types.IsNil(b) || types.IsNil(b.ListStrategyGroupRequest) {
		return nil
	}
	return &bo.QueryStrategyGroupListParams{
		Keyword: b.ListStrategyGroupRequest.GetKeyword(),
		TeamID:  b.ListStrategyGroupRequest.GetTeamId(),
		Status:  vobj.Status(b.ListStrategyGroupRequest.GetStatus()),
		Page:    types.NewPagination(b.ListStrategyGroupRequest.GetPagination()),
	}
}
