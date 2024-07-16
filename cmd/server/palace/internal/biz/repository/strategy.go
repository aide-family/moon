package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type Strategy interface {
	// CreateStrategy 创建策略
	CreateStrategy(ctx context.Context, params *bo.CreateStrategyParams) (*bizmodel.Strategy, error)

	// UpdateByID 更新策略
	UpdateByID(ctx context.Context, params *bo.UpdateStrategyParams) error

	// GetByID 获取策略详情
	GetByID(ctx context.Context, params *bo.GetStrategyDetailParams) (*bizmodel.Strategy, error)

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, params *bo.UpdateStrategyStatusParams) error

	// FindByPage 策略分页列表
	FindByPage(ctx context.Context, params *bo.QueryStrategyListParams) ([]*bizmodel.Strategy, error)

	// DeleteByID 删除策略
	DeleteByID(ctx context.Context, params *bo.DelStrategyParams) error
}
