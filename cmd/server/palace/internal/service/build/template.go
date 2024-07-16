package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

type TemplateStrategyBuilder struct {
	*model.StrategyTemplate
}

func NewTemplateStrategyBuilder(templateStrategy *model.StrategyTemplate) *TemplateStrategyBuilder {
	return &TemplateStrategyBuilder{
		StrategyTemplate: templateStrategy,
	}
}

func (b *TemplateStrategyBuilder) ToApi(ctx context.Context) *admin.StrategyTemplate {
	if types.IsNil(b) || types.IsNil(b.StrategyTemplate) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	return &admin.StrategyTemplate{
		Id:    b.ID,
		Alert: b.Alert,
		Expr:  b.Expr,
		Levels: types.SliceTo(b.StrategyLevelTemplates, func(item *model.StrategyLevelTemplate) *admin.StrategyLevelTemplate {
			return NewTemplateStrategyLevelBuilder(item).ToApi()
		}),
		Labels:      b.Labels.Map(),
		Annotations: b.Annotations,
		Status:      api.Status(b.Status),
		CreatedAt:   b.CreatedAt.String(),
		UpdatedAt:   b.UpdatedAt.String(),
		Remark:      b.Remark,
		Creator:     NewUserBuilder(cache.GetUser(ctx, b.CreatorID)).ToApi(),
		Categories: types.SliceTo(b.Categories, func(item *model.SysDict) *admin.Select {
			return NewDictBuild(item).ToApiSelect()
		}),
	}
}

type TemplateStrategyLevelBuilder struct {
	*model.StrategyLevelTemplate
}

func NewTemplateStrategyLevelBuilder(level *model.StrategyLevelTemplate) *TemplateStrategyLevelBuilder {
	return &TemplateStrategyLevelBuilder{
		StrategyLevelTemplate: level,
	}
}

func (b *TemplateStrategyLevelBuilder) ToApi() *admin.StrategyLevelTemplate {
	if types.IsNil(b) || types.IsNil(b.StrategyLevelTemplate) {
		return nil
	}
	return &admin.StrategyLevelTemplate{
		Id:          b.ID,
		Duration:    b.Duration.GetDuration(),
		Count:       b.Count,
		SustainType: api.SustainType(b.SustainType),
		Status:      api.Status(b.Status),
		LevelId:     b.LevelID,
		Level:       NewDictBuild(b.Level).ToApiSelect(),
		Threshold:   b.Threshold,
		StrategyId:  b.StrategyTemplateID,
		Condition:   api.Condition(b.Condition),
	}
}
