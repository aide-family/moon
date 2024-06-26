package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

// Template .
type Template interface {
	// CreateTemplateStrategy 创建模板策略
	CreateTemplateStrategy(ctx context.Context, templateStrategy *bo.CreateTemplateStrategyParams) error

	// UpdateTemplateStrategy 更新模板策略
	UpdateTemplateStrategy(ctx context.Context, templateStrategy *bo.UpdateTemplateStrategyParams) error

	// DeleteTemplateStrategy 删除模板策略
	DeleteTemplateStrategy(ctx context.Context, id uint32) error

	// GetTemplateStrategy 获取模板策略
	GetTemplateStrategy(ctx context.Context, id uint32) (*model.StrategyTemplate, error)

	// ListTemplateStrategy 获取模板策略列表
	ListTemplateStrategy(ctx context.Context, params *bo.QueryTemplateStrategyListParams) ([]*model.StrategyTemplate, error)

	// UpdateTemplateStrategyStatus 更新模板策略状态
	UpdateTemplateStrategyStatus(ctx context.Context, status vobj.Status, ids ...uint32) error
}
