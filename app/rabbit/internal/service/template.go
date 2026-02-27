package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/biz/bo"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func NewTemplateService(templateBiz *biz.Template) *TemplateService {
	return &TemplateService{
		templateBiz: templateBiz,
	}
}

type TemplateService struct {
	apiv1.UnimplementedTemplateServer
	templateBiz *biz.Template
}

func (s *TemplateService) CreateTemplate(ctx context.Context, req *apiv1.CreateTemplateRequest) (*apiv1.CreateTemplateReply, error) {
	createBo, err := bo.NewCreateTemplateBo(req)
	if err != nil {
		return nil, err
	}
	uid, err := s.templateBiz.CreateTemplate(ctx, createBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateTemplateReply{Uid: uid.Int64()}, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, req *apiv1.UpdateTemplateRequest) (*apiv1.UpdateTemplateReply, error) {
	updateBo, err := bo.NewUpdateTemplateBo(req)
	if err != nil {
		return nil, err
	}
	if err := s.templateBiz.UpdateTemplate(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateTemplateReply{}, nil
}

func (s *TemplateService) UpdateTemplateStatus(ctx context.Context, req *apiv1.UpdateTemplateStatusRequest) (*apiv1.UpdateTemplateStatusReply, error) {
	updateBo := bo.NewUpdateTemplateStatusBo(req)
	if err := s.templateBiz.UpdateTemplateStatus(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateTemplateStatusReply{}, nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, req *apiv1.DeleteTemplateRequest) (*apiv1.DeleteTemplateReply, error) {
	if err := s.templateBiz.DeleteTemplate(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.DeleteTemplateReply{}, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, req *apiv1.GetTemplateRequest) (*apiv1.TemplateItem, error) {
	templateBo, err := s.templateBiz.GetTemplate(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return templateBo.ToAPIV1TemplateItem(), nil
}

func (s *TemplateService) ListTemplate(ctx context.Context, req *apiv1.ListTemplateRequest) (*apiv1.ListTemplateReply, error) {
	listBo := bo.NewListTemplateBo(req)
	pageResponseBo, err := s.templateBiz.ListTemplate(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListTemplateReply(pageResponseBo), nil
}

func (s *TemplateService) SelectTemplate(ctx context.Context, req *apiv1.SelectTemplateRequest) (*apiv1.SelectTemplateReply, error) {
	selectBo := bo.NewSelectTemplateBo(req)
	result, err := s.templateBiz.SelectTemplate(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectTemplateReply(&bo.SelectTemplateReplyParams{
		Items:   result.Items,
		Total:   result.Total,
		LastUID: result.LastUID,
		Limit:   req.Limit,
	}), nil
}
