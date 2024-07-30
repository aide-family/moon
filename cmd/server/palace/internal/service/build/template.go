package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	templateapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// TemplateModelBuilder 模版模型构造器
	TemplateModelBuilder interface {
		ToAPI(ctx context.Context) *admin.StrategyTemplate
	}

	// TemplateRequestBuilder 模版请求参数构造器
	TemplateRequestBuilder interface {
		ToCreateTemplateBO() *bo.CreateTemplateStrategyParams
		ToUpdateTemplateBO() *bo.UpdateTemplateStrategyParams
	}

	templateStrategyBuilder struct {
		// model
		*model.StrategyTemplate

		// request
		CreateStrategy *templateapi.CreateTemplateStrategyRequest
		UpdateStrategy *templateapi.UpdateTemplateStrategyRequest

		// context
		ctx context.Context
	}

	// TemplateLevelBuilder 模版级别构造器
	TemplateLevelBuilder interface {
		ToAPI() *admin.StrategyLevelTemplate
	}

	templateStrategyLevelBuilder struct {
		*model.StrategyLevelTemplate
		ctx context.Context
	}
)

func (b *templateStrategyBuilder) ToAPI(ctx context.Context) *admin.StrategyTemplate {
	if types.IsNil(b) || types.IsNil(b.StrategyTemplate) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	return &admin.StrategyTemplate{
		Id:    b.ID,
		Alert: b.Alert,
		Expr:  b.Expr,
		Levels: types.SliceTo(b.StrategyLevelTemplates, func(item *model.StrategyLevelTemplate) *admin.StrategyLevelTemplate {
			return NewBuilder().WithAPITemplateStrategyLevel(item).ToAPI()
		}),
		Labels:      b.Labels.Map(),
		Annotations: b.Annotations,
		Status:      api.Status(b.Status),
		CreatedAt:   b.CreatedAt.String(),
		UpdatedAt:   b.UpdatedAt.String(),
		Remark:      b.Remark,
		Creator:     NewBuilder().WithAPIUserBo(cache.GetUser(ctx, b.CreatorID)).ToAPI(),
		Categories: types.SliceTo(b.Categories, func(item *model.SysDict) *admin.SelectItem {
			return NewBuilder().WithAPIDictSelect(item).ToAPISelect()
		}),
	}
}

func (b *templateStrategyBuilder) ToCreateTemplateBO() *bo.CreateTemplateStrategyParams {
	if types.IsNil(b) || types.IsNil(b.CreateStrategy) {
		return nil
	}
	strategyLevelTemplates := make([]*bo.CreateStrategyLevelTemplate, 0, len(b.CreateStrategy.GetLevel()))
	for levelID, mutationStrategyLevelTemplate := range b.CreateStrategy.GetLevel() {
		strategyLevelTemplates = append(strategyLevelTemplates, &bo.CreateStrategyLevelTemplate{
			Duration:    &types.Duration{Duration: mutationStrategyLevelTemplate.Duration},
			Count:       mutationStrategyLevelTemplate.GetCount(),
			SustainType: vobj.Sustain(mutationStrategyLevelTemplate.GetSustainType()),
			Condition:   vobj.Condition(mutationStrategyLevelTemplate.GetCondition()),
			Threshold:   mutationStrategyLevelTemplate.GetThreshold(),
			LevelID:     levelID,
			Status:      vobj.StatusEnable,
		})
	}

	return &bo.CreateTemplateStrategyParams{
		Alert:                  b.CreateStrategy.GetAlert(),
		Expr:                   b.CreateStrategy.GetExpr(),
		Status:                 vobj.StatusEnable,
		Remark:                 b.CreateStrategy.GetRemark(),
		Labels:                 vobj.NewLabels(b.CreateStrategy.GetLabels()),
		Annotations:            b.CreateStrategy.GetAnnotations(),
		StrategyLevelTemplates: strategyLevelTemplates,
		CategoriesIDs:          b.CreateStrategy.GetCategoriesIds(),
	}
}

func (b *templateStrategyBuilder) ToUpdateTemplateBO() *bo.UpdateTemplateStrategyParams {
	if types.IsNil(b) || types.IsNil(b.UpdateStrategy) {
		return nil
	}
	strategyLevelTemplates := make([]*bo.CreateStrategyLevelTemplate, 0, len(b.UpdateStrategy.GetLevel()))
	for levelID, mutationStrategyLevelTemplate := range b.UpdateStrategy.GetLevel() {
		strategyLevelTemplates = append(strategyLevelTemplates, &bo.CreateStrategyLevelTemplate{
			StrategyTemplateID: b.UpdateStrategy.GetId(),
			Duration:           &types.Duration{Duration: mutationStrategyLevelTemplate.Duration},
			Count:              mutationStrategyLevelTemplate.GetCount(),
			SustainType:        vobj.Sustain(mutationStrategyLevelTemplate.GetSustainType()),
			Condition:          vobj.Condition(mutationStrategyLevelTemplate.GetCondition()),
			Threshold:          mutationStrategyLevelTemplate.GetThreshold(),
			LevelID:            levelID,
			Status:             vobj.StatusEnable,
		})
	}
	return &bo.UpdateTemplateStrategyParams{
		Data: bo.CreateTemplateStrategyParams{
			Alert:                  b.UpdateStrategy.GetAlert(),
			Expr:                   b.UpdateStrategy.GetExpr(),
			Status:                 vobj.StatusEnable,
			Remark:                 b.UpdateStrategy.GetRemark(),
			Labels:                 vobj.NewLabels(b.UpdateStrategy.GetLabels()),
			Annotations:            b.UpdateStrategy.GetAnnotations(),
			StrategyLevelTemplates: strategyLevelTemplates,
		},
		ID: b.UpdateStrategy.Id,
	}
}

func (b *templateStrategyLevelBuilder) ToAPI() *admin.StrategyLevelTemplate {
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
		Level:       NewBuilder().WithAPIDictSelect(b.Level).ToAPISelect(),
		Threshold:   b.Threshold,
		StrategyId:  b.StrategyTemplateID,
		Condition:   api.Condition(b.Condition),
	}
}
