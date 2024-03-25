package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type NotifyTemplateBiz struct {
	log *log.Helper

	repository.NotifyTemplateRepo
}

func NewNotifyTemplateBiz(
	repo repository.NotifyTemplateRepo,
	logger log.Logger,
) *NotifyTemplateBiz {
	return &NotifyTemplateBiz{
		log:                log.NewHelper(log.With(logger, "module", "biz.notify_template")),
		NotifyTemplateRepo: repo,
	}
}

// CreateTemplate 创建通知模板
func (n *NotifyTemplateBiz) CreateTemplate(ctx context.Context, req *bo.NotifyTemplateCreateBO) (*bo.NotifyTemplateBO, error) {
	createParams := &bo.NotifyTemplateBO{
		Content:    req.Content,
		StrategyID: req.StrategyID,
		NotifyType: req.NotifyType,
	}
	newNotifyTemplate, err := n.NotifyTemplateRepo.Create(ctx, createParams)
	if err != nil {
		return nil, err
	}
	return newNotifyTemplate, nil
}

// UpdateTemplate 更新通知模板
func (n *NotifyTemplateBiz) UpdateTemplate(ctx context.Context, req *bo.NotifyTemplateUpdateBo) error {
	updateParams, err := n.NotifyTemplateRepo.Get(ctx, basescopes.IdGT(req.Id))
	if err != nil {
		return err
	}
	updateParams.Content = req.Content
	updateParams.StrategyID = req.StrategyID
	updateParams.NotifyType = req.NotifyType

	if err = n.NotifyTemplateRepo.Update(ctx, updateParams); err != nil {
		return err
	}
	return nil
}

// DeleteTemplate 删除通知模板
func (n *NotifyTemplateBiz) DeleteTemplate(ctx context.Context, templateId uint32) error {
	return n.NotifyTemplateRepo.Delete(ctx, basescopes.IdGT(templateId))
}

// GetTemplate 获取通知模板详情
func (n *NotifyTemplateBiz) GetTemplate(ctx context.Context, templateId uint32) (*bo.NotifyTemplateBO, error) {
	return n.NotifyTemplateRepo.Get(ctx, basescopes.IdGT(templateId))
}

// ListTemplate 获取通知模板列表
func (n *NotifyTemplateBiz) ListTemplate(ctx context.Context, req *bo.NotifyTemplateListBo) ([]*bo.NotifyTemplateBO, error) {
	return n.NotifyTemplateRepo.List(ctx, req.Page, do.PromStrategyNotifyTemplateWhereStrategyID(req.StrategyId))
}
