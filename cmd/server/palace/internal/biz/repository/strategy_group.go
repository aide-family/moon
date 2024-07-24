package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type StrategyGroup interface {

	// CreateStrategyGroup 创建策略组
	CreateStrategyGroup(ctx context.Context, params *bo.CreateStrategyGroupParams) (*bizmodel.StrategyGroup, error)
	// UpdateStrategyGroup 更新策略组
	UpdateStrategyGroup(ctx context.Context, params *bo.UpdateStrategyGroupParams) error
	// DeleteStrategyGroup 删除策略组
	DeleteStrategyGroup(ctx context.Context, params *bo.DelStrategyGroupParams) error
	// GetStrategyGroup 获取策略详情
	GetStrategyGroup(ctx context.Context, params *bo.GetStrategyGroupDetailParams) (*bizmodel.StrategyGroup, error)
	// StrategyGroupPage 策略列表
	StrategyGroupPage(ctx context.Context, params *bo.QueryStrategyGroupListParams) ([]*bizmodel.StrategyGroup, error)
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, params *bo.UpdateStrategyGroupStatusParams) error
}
