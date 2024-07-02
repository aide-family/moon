package strategy

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type TemplateService struct {
	strategyapi.UnimplementedTemplateServer

	templateBiz *biz.TemplateBiz
}

func NewTemplateService(templateBiz *biz.TemplateBiz) *TemplateService {
	return &TemplateService{
		templateBiz: templateBiz,
	}
}

func (s *TemplateService) CreateTemplateStrategy(ctx context.Context, req *strategyapi.CreateTemplateStrategyRequest) (*strategyapi.CreateTemplateStrategyReply, error) {
	strategyLevelTemplates := make([]*bo.CreateStrategyLevelTemplate, 0, len(req.GetLevel()))
	for levelID, mutationStrategyLevelTemplate := range req.GetLevel() {
		strategyLevelTemplates = append(strategyLevelTemplates, &bo.CreateStrategyLevelTemplate{
			Duration:    &types.Duration{Duration: mutationStrategyLevelTemplate.Duration},
			Count:       mutationStrategyLevelTemplate.GetCount(),
			SustainType: vobj.Sustain(mutationStrategyLevelTemplate.GetSustainType()),
			Interval:    &types.Duration{Duration: mutationStrategyLevelTemplate.Interval},
			Condition:   mutationStrategyLevelTemplate.GetCondition(),
			Threshold:   mutationStrategyLevelTemplate.GetThreshold(),
			LevelID:     levelID,
			Status:      vobj.StatusEnable,
		})
	}

	params := &bo.CreateTemplateStrategyParams{
		Alert:                  req.GetAlert(),
		Expr:                   req.GetExpr(),
		Status:                 vobj.StatusEnable,
		Remark:                 req.GetRemark(),
		Labels:                 req.GetLabels(),
		Annotations:            req.GetAnnotations(),
		StrategyLevelTemplates: strategyLevelTemplates,
	}
	if err := s.templateBiz.CreateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.CreateTemplateStrategyReply{}, nil
}

func (s *TemplateService) UpdateTemplateStrategy(ctx context.Context, req *strategyapi.UpdateTemplateStrategyRequest) (*strategyapi.UpdateTemplateStrategyReply, error) {
	strategyLevelTemplates := make([]*bo.CreateStrategyLevelTemplate, 0, len(req.GetLevel()))
	for levelID, mutationStrategyLevelTemplate := range req.GetLevel() {
		strategyLevelTemplates = append(strategyLevelTemplates, &bo.CreateStrategyLevelTemplate{
			StrategyTemplateID: req.GetId(),
			Duration:           &types.Duration{Duration: mutationStrategyLevelTemplate.Duration},
			Count:              mutationStrategyLevelTemplate.GetCount(),
			SustainType:        vobj.Sustain(mutationStrategyLevelTemplate.GetSustainType()),
			Interval:           &types.Duration{Duration: mutationStrategyLevelTemplate.Interval},
			Condition:          mutationStrategyLevelTemplate.GetCondition(),
			Threshold:          mutationStrategyLevelTemplate.GetThreshold(),
			LevelID:            levelID,
			Status:             vobj.StatusEnable,
		})
	}
	params := &bo.UpdateTemplateStrategyParams{
		Data: bo.CreateTemplateStrategyParams{
			Alert:                  req.GetAlert(),
			Expr:                   req.GetExpr(),
			Status:                 vobj.StatusEnable,
			Remark:                 req.GetRemark(),
			Labels:                 req.GetLabels(),
			Annotations:            req.GetAnnotations(),
			StrategyLevelTemplates: strategyLevelTemplates,
		},
		ID: req.Id,
	}
	if err := s.templateBiz.UpdateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateTemplateStrategyReply{}, nil
}

func (s *TemplateService) DeleteTemplateStrategy(ctx context.Context, req *strategyapi.DeleteTemplateStrategyRequest) (*strategyapi.DeleteTemplateStrategyReply, error) {
	if err := s.templateBiz.DeleteTemplateStrategy(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &strategyapi.DeleteTemplateStrategyReply{}, nil
}

func (s *TemplateService) GetTemplateStrategy(ctx context.Context, req *strategyapi.GetTemplateStrategyRequest) (*strategyapi.GetTemplateStrategyReply, error) {
	detail, err := s.templateBiz.GetTemplateStrategy(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &strategyapi.GetTemplateStrategyReply{
		Detail: build.NewTemplateStrategyBuilder(detail).ToApi(),
	}, nil
}

func (s *TemplateService) ListTemplateStrategy(ctx context.Context, req *strategyapi.ListTemplateStrategyRequest) (*strategyapi.ListTemplateStrategyReply, error) {
	params := &bo.QueryTemplateStrategyListParams{
		Page:   types.NewPagination(req.GetPagination()),
		Alert:  req.GetKeyword(),
		Status: vobj.Status(req.GetStatus()),
	}
	list, err := s.templateBiz.ListTemplateStrategy(ctx, params)
	if err != nil {
		return nil, err
	}
	return &strategyapi.ListTemplateStrategyReply{
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
		List: types.SliceTo(list, func(item *model.StrategyTemplate) *admin.StrategyTemplate {
			return build.NewTemplateStrategyBuilder(item).ToApi()
		}),
	}, nil
}

func (s *TemplateService) UpdateTemplateStrategyStatus(ctx context.Context, req *strategyapi.UpdateTemplateStrategyStatusRequest) (*strategyapi.UpdateTemplateStrategyStatusReply, error) {
	if err := s.templateBiz.UpdateTemplateStrategyStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()...); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateTemplateStrategyStatusReply{}, nil
}
