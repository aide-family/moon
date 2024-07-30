package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewTemplateBiz 创建模板管理业务
func NewTemplateBiz(templateRepository repository.Template) *TemplateBiz {
	return &TemplateBiz{
		templateRepository: templateRepository,
	}
}

// TemplateBiz 模板管理业务
type TemplateBiz struct {
	templateRepository repository.Template
}

// CreateTemplateStrategy 创建模板策略
func (b *TemplateBiz) CreateTemplateStrategy(ctx context.Context, templateStrategy *bo.CreateTemplateStrategyParams) error {
	return b.templateRepository.CreateTemplateStrategy(ctx, templateStrategy)
}

// UpdateTemplateStrategy 更新模板策略
func (b *TemplateBiz) UpdateTemplateStrategy(ctx context.Context, templateStrategy *bo.UpdateTemplateStrategyParams) error {
	if types.IsNil(templateStrategy.Data) {
		return merr.ErrorI18nSystemErr(ctx)
	}
	// 查询策略详情
	_, err := b.templateRepository.GetTemplateStrategy(ctx, templateStrategy.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nStrategyTemplateNotFoundErr(ctx)
		}
		return err
	}
	return b.templateRepository.UpdateTemplateStrategy(ctx, templateStrategy)
}

// DeleteTemplateStrategy 删除模板策略
func (b *TemplateBiz) DeleteTemplateStrategy(ctx context.Context, id uint32) error {
	// 查询策略详情
	_, err := b.templateRepository.GetTemplateStrategy(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nStrategyTemplateNotFoundErr(ctx)
		}
		return err
	}
	return b.templateRepository.DeleteTemplateStrategy(ctx, id)
}

// GetTemplateStrategy 获取模板策略详情
func (b *TemplateBiz) GetTemplateStrategy(ctx context.Context, id uint32) (*model.StrategyTemplate, error) {
	return b.templateRepository.GetTemplateStrategy(ctx, id)
}

// ListTemplateStrategy 获取模板策略列表
func (b *TemplateBiz) ListTemplateStrategy(ctx context.Context, params *bo.QueryTemplateStrategyListParams) ([]*model.StrategyTemplate, error) {
	return b.templateRepository.ListTemplateStrategy(ctx, params)
}

// UpdateTemplateStrategyStatus 更新模板策略状态
func (b *TemplateBiz) UpdateTemplateStrategyStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	return b.templateRepository.UpdateTemplateStrategyStatus(ctx, status, ids...)
}
