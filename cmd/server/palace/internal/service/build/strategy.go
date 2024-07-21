package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	StrategyModelBuilder interface {
		ToApi(ctx context.Context) *admin.Strategy
	}

	StrategyRequestBuilder interface {
		ToCreateStrategyBO() *bo.CreateStrategyParams
		ToUpdateStrategyBO() *bo.UpdateStrategyParams
	}

	StrategyLevelModelBuilder interface {
		ToApi() *admin.StrategyLevel
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
)

// ToApi 转换为API层数据
func (b *strategyBuilder) ToApi(ctx context.Context) *admin.Strategy {
	if types.IsNil(b) || types.IsNil(b.Strategy) {
		return nil
	}
	strategyLevels := types.SliceToWithFilter(b.Strategy.StrategyLevel, func(level *bizmodel.StrategyLevel) (*admin.StrategyLevel, bool) {
		return NewBuilder().WithApiStrategyLevel(level).ToApi(), true
	})

	return &admin.Strategy{
		Name:        b.Strategy.Name,
		Id:          b.Strategy.ID,
		Expr:        b.Strategy.Expr,
		Labels:      b.Strategy.Labels.Map(),
		Annotations: b.Strategy.Annotations,
		Datasource: types.SliceTo(b.Strategy.Datasource, func(datasource *bizmodel.Datasource) *admin.Datasource {
			return NewBuilder().WithContext(ctx).WithDoDatasource(datasource).ToApi()
		}),
		StrategyTemplateId: b.Strategy.StrategyTemplateID,
		Levels:             strategyLevels,
		Status:             api.Status(b.Strategy.Status),
		Step:               b.Strategy.Step,
		SourceType:         api.TemplateSourceType(b.Strategy.StrategyTemplateSource),
	}
}

func (b *strategyBuilder) ToCreateStrategyBO() *bo.CreateStrategyParams {
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
		TemplateId:    b.CreateStrategy.GetTemplateId(),
		GroupId:       b.CreateStrategy.GetGroupId(),
		Name:          b.CreateStrategy.GetName(),
		Remark:        b.CreateStrategy.GetRemark(),
		Status:        vobj.Status(b.CreateStrategy.GetStatus()),
		Step:          b.CreateStrategy.GetStep(),
		SourceType:    vobj.TemplateSourceType(b.CreateStrategy.GetSourceType()),
		DatasourceIds: b.CreateStrategy.GetDatasourceIds(),
		Labels:        vobj.NewLabels(b.CreateStrategy.GetLabels()),
		Annotations:   b.CreateStrategy.GetAnnotations(),
		StrategyLevel: strategyLevels,
	}
}

func (b *strategyBuilder) ToUpdateStrategyBO() *bo.UpdateStrategyParams {

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
			TemplateId:    b.UpdateStrategy.GetData().GetTemplateId(),
			GroupId:       b.UpdateStrategy.GetData().GetGroupId(),
			Name:          b.UpdateStrategy.GetData().GetName(),
			Remark:        b.UpdateStrategy.GetData().GetRemark(),
			Status:        vobj.Status(b.UpdateStrategy.GetData().GetStatus()),
			Step:          b.UpdateStrategy.GetData().GetStep(),
			SourceType:    vobj.TemplateSourceType(b.UpdateStrategy.GetData().GetSourceType()),
			DatasourceIds: b.UpdateStrategy.GetData().GetDatasourceIds(),
			StrategyLevel: strategyLevels,
		},
	}
}

func (b *strategyLevelBuilder) ToApi() *admin.StrategyLevel {
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
