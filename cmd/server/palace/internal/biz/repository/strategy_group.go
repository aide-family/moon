package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	// StrategyGroup 策略组接口
	StrategyGroup interface {
		// CreateStrategyGroup 创建策略组
		CreateStrategyGroup(context.Context, *bo.CreateStrategyGroupParams) (*bizmodel.StrategyGroup, error)
		// UpdateStrategyGroup 更新策略组
		UpdateStrategyGroup(context.Context, *bo.UpdateStrategyGroupParams) error
		// DeleteStrategyGroup 删除策略组
		DeleteStrategyGroup(context.Context, *bo.DelStrategyGroupParams) error
		// GetStrategyGroup 获取策略详情
		GetStrategyGroup(context.Context, uint32) (*bizmodel.StrategyGroup, error)
		// StrategyGroupPage 策略列表
		StrategyGroupPage(context.Context, *bo.QueryStrategyGroupListParams) ([]*bizmodel.StrategyGroup, error)
		// UpdateStatus 更新状态
		UpdateStatus(context.Context, *bo.UpdateStrategyGroupStatusParams) error
	}

	// StrategyCountRepo 策略数量接口
	StrategyCountRepo interface {
		// FindStrategyCount 获取策略总数
		FindStrategyCount(context.Context, *bo.GetStrategyCountParams) ([]*bo.StrategyCountModel, error)
	}
)
