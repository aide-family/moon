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
	strategyLevelTemplates := make([]*model.StrategyLevelTemplate, 0, len(req.GetLevel()))
	for levelID, mutationStrategyLevelTemplate := range req.GetLevel() {
		strategyLevelTemplates = append(strategyLevelTemplates, &model.StrategyLevelTemplate{
			StrategyID:  0,
			Duration:    *mutationStrategyLevelTemplate.GetDuration(),
			Count:       mutationStrategyLevelTemplate.GetCount(),
			SustainType: vobj.Sustain(mutationStrategyLevelTemplate.SustainType),
			Interval:    *mutationStrategyLevelTemplate.GetInterval(),
			Condition:   mutationStrategyLevelTemplate.Condition,
			Threshold:   mutationStrategyLevelTemplate.GetThreshold(),
			LevelID:     levelID,
			Status:      vobj.StatusEnable,
		})
	}
	params := &bo.CreateTemplateStrategyParams{
		StrategyTemplate: &model.StrategyTemplate{
			Alert:                  req.GetAlert(),
			Expr:                   req.GetExpr(),
			Status:                 vobj.StatusEnable,
			Remark:                 req.GetRemark(),
			Labels:                 req.GetLabels(),
			Annotations:            req.GetAnnotations(),
			StrategyLevelTemplates: strategyLevelTemplates,
		},
	}
	if err := s.templateBiz.CreateTemplateStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.CreateTemplateStrategyReply{}, nil
}

func (s *TemplateService) UpdateTemplateStrategy(ctx context.Context, req *strategyapi.UpdateTemplateStrategyRequest) (*strategyapi.UpdateTemplateStrategyReply, error) {
	strategyLevelTemplates := make([]*model.StrategyLevelTemplate, 0, len(req.GetLevel()))
	for levelID, mutationStrategyLevelTemplate := range req.GetLevel() {
		strategyLevelTemplates = append(strategyLevelTemplates, &model.StrategyLevelTemplate{
			StrategyID:  req.GetId(),
			Duration:    *mutationStrategyLevelTemplate.GetDuration(),
			Count:       mutationStrategyLevelTemplate.GetCount(),
			SustainType: vobj.Sustain(mutationStrategyLevelTemplate.SustainType),
			Interval:    *mutationStrategyLevelTemplate.GetInterval(),
			Condition:   mutationStrategyLevelTemplate.Condition,
			Threshold:   mutationStrategyLevelTemplate.GetThreshold(),
			LevelID:     levelID,
			Status:      vobj.StatusEnable,
		})
	}
	params := &bo.UpdateTemplateStrategyParams{
		StrategyTemplate: &model.StrategyTemplate{
			AllFieldModel: model.AllFieldModel{
				BaseModel: model.BaseModel{ID: req.GetId()},
			},
			Alert:                  req.GetAlert(),
			Expr:                   req.GetExpr(),
			Status:                 vobj.StatusEnable,
			Remark:                 req.GetRemark(),
			Labels:                 req.GetLabels(),
			Annotations:            req.GetAnnotations(),
			StrategyLevelTemplates: strategyLevelTemplates,
		},
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
		Page:   types.NewPage(int(req.GetPageNum()), int(req.GetPageSize())),
		Alert:  req.GetAlert(),
		Status: vobj.Status(req.GetStatus()),
	}
	list, err := s.templateBiz.ListTemplateStrategy(ctx, params)
	if err != nil {
		return nil, err
	}
	return &strategyapi.ListTemplateStrategyReply{
		Total: int64(params.Page.GetTotal()),
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
