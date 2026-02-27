package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/biz/bo"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func NewWebhookService(webhookConfigBiz *biz.WebhookConfig) *WebhookService {
	return &WebhookService{
		webhookConfigBiz: webhookConfigBiz,
	}
}

type WebhookService struct {
	apiv1.UnimplementedWebhookServer
	webhookConfigBiz *biz.WebhookConfig
}

func (s *WebhookService) CreateWebhook(ctx context.Context, req *apiv1.CreateWebhookRequest) (*apiv1.CreateWebhookReply, error) {
	createBo := bo.NewCreateWebhookBo(req)
	uid, err := s.webhookConfigBiz.CreateWebhook(ctx, createBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateWebhookReply{Uid: uid.Int64()}, nil
}

func (s *WebhookService) UpdateWebhook(ctx context.Context, req *apiv1.UpdateWebhookRequest) (*apiv1.UpdateWebhookReply, error) {
	updateBo := bo.NewUpdateWebhookBo(req)
	if err := s.webhookConfigBiz.UpdateWebhook(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateWebhookReply{}, nil
}

func (s *WebhookService) UpdateWebhookStatus(ctx context.Context, req *apiv1.UpdateWebhookStatusRequest) (*apiv1.UpdateWebhookStatusReply, error) {
	updateBo := bo.NewUpdateWebhookStatusBo(req)
	if err := s.webhookConfigBiz.UpdateWebhookStatus(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateWebhookStatusReply{}, nil
}

func (s *WebhookService) DeleteWebhook(ctx context.Context, req *apiv1.DeleteWebhookRequest) (*apiv1.DeleteWebhookReply, error) {
	if err := s.webhookConfigBiz.DeleteWebhook(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &apiv1.DeleteWebhookReply{}, nil
}

func (s *WebhookService) GetWebhook(ctx context.Context, req *apiv1.GetWebhookRequest) (*apiv1.WebhookItem, error) {
	webhookBo, err := s.webhookConfigBiz.GetWebhook(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return webhookBo.ToAPIV1WebhookItem(), nil
}

func (s *WebhookService) ListWebhook(ctx context.Context, req *apiv1.ListWebhookRequest) (*apiv1.ListWebhookReply, error) {
	listBo := bo.NewListWebhookBo(req)
	pageResponseBo, err := s.webhookConfigBiz.ListWebhook(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListWebhookReply(pageResponseBo), nil
}

func (s *WebhookService) SelectWebhook(ctx context.Context, req *apiv1.SelectWebhookRequest) (*apiv1.SelectWebhookReply, error) {
	selectBo := bo.NewSelectWebhookBo(req)
	result, err := s.webhookConfigBiz.SelectWebhook(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectWebhookReply(&bo.SelectWebhookReplyParams{
		Items:   result.Items,
		Total:   result.Total,
		LastUID: result.LastUID,
		Limit:   req.Limit,
	}), nil
}
