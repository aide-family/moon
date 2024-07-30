package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// Strategy 策略管理接口
type Strategy interface {
	// CreateStrategy 创建策略
	CreateStrategy(context.Context, *bo.CreateStrategyParams) (*bizmodel.Strategy, error)

	// UpdateByID 更新策略
	UpdateByID(context.Context, *bo.UpdateStrategyParams) error

	// GetByID 获取策略详情
	GetByID(context.Context, *bo.GetStrategyDetailParams) (*bizmodel.Strategy, error)

	// UpdateStatus 更新状态
	UpdateStatus(context.Context, *bo.UpdateStrategyStatusParams) error

	// FindByPage 策略分页列表
	FindByPage(context.Context, *bo.QueryStrategyListParams) ([]*bizmodel.Strategy, error)

	// DeleteByID 删除策略
	DeleteByID(context.Context, *bo.DelStrategyParams) error

	// CopyStrategy 复制策略
	CopyStrategy(context.Context, *bo.CopyStrategyParams) (*bizmodel.Strategy, error)
}
